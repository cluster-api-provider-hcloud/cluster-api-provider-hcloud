package server

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubeletv1beta1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/kubelet/v1beta1"
	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/server/userdata"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/utils"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/record"
)

type Service struct {
	scope *scope.MachineScope
}

func NewService(scope *scope.MachineScope) *Service {
	return &Service{
		scope: scope,
	}
}

var errNotImplemented = errors.New("Not implemented")

const etcdMountPath = "/var/lib/etcd"

func stringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *Service) genericLabels() map[string]string {
	return map[string]string{
		infrav1.ClusterTagKey(s.scope.HcloudCluster.Name): string(infrav1.ResourceLifecycleOwned),
	}
}

func (s *Service) createLabels() map[string]string {
	m := s.genericLabels()
	m[infrav1.MachineNameTagKey] = s.scope.Name()
	var machineType string
	if s.scope.IsControlPlane() == true {
		machineType = "control_plane"
	} else {
		machineType = "worker"
	}
	m["machine_type"] = machineType
	return m
}

func (s *Service) labels() map[string]string {
	m := s.genericLabels()
	m[infrav1.MachineNameTagKey] = s.scope.Name()
	return m
}

type intSlice []int

func (s intSlice) contains(e int) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

func (s *Service) addServerToLoadBalancer(server *hcloud.Server) error {

	if len(s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers) == 0 {
		s.scope.V(2).Info("No load balancer found for adding server as target", "Server", server.ID)
		return nil
	}

	var targetIDs intSlice
	targetIDs = s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers[0].Targets

	if targetIDs.contains(server.ID) {
		return nil
	}

	myBool := true

	loadBalancerAddServerTargetOpts := hcloud.LoadBalancerAddServerTargetOpts{Server: server, UsePrivateIP: &myBool}

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(context.Background(), hcloud.LoadBalancerListOpts{})
	if err != nil {
		return err
	}
	// This only works if there is only one load balancer
	lb := loadBalancers[0]
	_, _, err = s.scope.HcloudClient().AddTargetServerToLoadBalancer(context.Background(), loadBalancerAddServerTargetOpts, lb)
	if err != nil {
		s.scope.V(2).Info("Could not add server as target to load balancer", "Server", server.ID, "Load Balancer", lb.ID)
		return err
	} else {
		record.Eventf(
			s.scope.HcloudCluster,
			"AddedAsTargetToLoadBalancer",
			"Added new server with id %d to the loadbalancer %v",
			server.ID, lb.ID)
	}
	return nil
}

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {
	// detect failure domain
	failureDomain, err := s.scope.GetFailureDomain()
	if err != nil {
		return nil, err
	}
	s.scope.HcloudMachine.Status.Location = infrav1.HcloudLocation(failureDomain)

	// gather image ID
	imageID, err := s.findImageIDBySpec(s.scope.Ctx, s.scope.HcloudMachine.Spec.Image)
	if err != nil {
		return nil, err
	}
	if imageID == nil {
		return &ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}
	if s.scope.HcloudMachine.Spec.Image == nil {
		s.scope.HcloudMachine.Spec.Image = &infrav1.HcloudImageSpec{}
	}
	s.scope.HcloudMachine.Spec.Image.ID = imageID
	s.scope.HcloudMachine.Status.ImageID = imageID

	// gather volumes
	volumes := make([]*hcloud.Volume, len(s.scope.HcloudMachine.Spec.Volumes))
	for pos, volume := range s.scope.HcloudMachine.Spec.Volumes {
		volumeObjectKey := types.NamespacedName{Namespace: s.scope.HcloudMachine.Namespace, Name: volume.VolumeRef}
		var hcloudVolume infrav1.HcloudVolume
		err := s.scope.Client.Get(
			ctx,
			volumeObjectKey,
			&hcloudVolume,
		)
		if apierrors.IsNotFound(err) {
			s.scope.V(1).Info("HcloudVolume is not found", "hcloudVolume", volumeObjectKey)
			return &reconcile.Result{}, nil
		} else if err != nil {
			return nil, err
		}
		if hcloudVolume.Status.VolumeID == nil {
			s.scope.V(1).Info("HcloudVolume is not existing yet", "hcloudVolume", volumeObjectKey)
			return &reconcile.Result{}, nil
		}
		volumes[pos] = &hcloud.Volume{
			ID: int(*hcloudVolume.Status.VolumeID),
		}
	}

	userDataInitial, err := s.scope.GetRawBootstrapData(ctx)
	if err == scope.ErrBootstrapDataNotReady {
		s.scope.V(1).Info("Bootstrap data is not ready yet")
		return &reconcile.Result{RequeueAfter: 15 * time.Second}, nil
	} else if err != nil {
		return nil, err
	}

	userData, err := userdata.NewFromReader(bytes.NewReader(userDataInitial))
	if err != nil {
		return nil, err
	}

	kubeadmConfig, err := userData.GetKubeadmConfig()
	if err != nil {
		return nil, err
	}

	cloudProviderKey := "cloud-provider"
	cloudProviderValue := "external"

	if s.scope.IsControlPlane() {
		// set up iptables proxy
		if kubeadmConfig.IsInit() {
			iptablesProxy, err := s.getIPTablesProxyFile()
			if err != nil {
				return nil, err
			}
			if err := userData.SetOrUpdateFile(iptablesProxy); err != nil {
				return nil, err
			}

			// enable TLS bootstrapping and rollover
			kubeadmConfig.KubeletConfiguration = &kubeletv1beta1.KubeletConfiguration{
				ServerTLSBootstrap: true,
				RotateCertificates: true,
			}

			if i := kubeadmConfig.InitConfiguration; i != nil {
				// set cloud provider external if nothing else is set
				if i.NodeRegistration.KubeletExtraArgs == nil {
					i.NodeRegistration.KubeletExtraArgs = make(map[string]string)
				}
				if _, ok := i.NodeRegistration.KubeletExtraArgs[cloudProviderKey]; !ok {
					i.NodeRegistration.KubeletExtraArgs[cloudProviderKey] = cloudProviderValue
				}
			} else {
				record.Warnf(
					s.scope.HcloudMachine,
					"UnexpectedUserData",
					"UserData for a control plane comes without a InitConfiguration",
				)
			}

			if c := kubeadmConfig.ClusterConfiguration; c != nil {
				if c.APIServer.ExtraArgs == nil {
					c.APIServer.ExtraArgs = make(map[string]string)
				}
				if c.ControllerManager.ExtraArgs == nil {
					c.ControllerManager.ExtraArgs = make(map[string]string)
				}

				// set cloud provider external if nothing is set
				if _, ok := c.APIServer.ExtraArgs[cloudProviderKey]; !ok {
					c.APIServer.ExtraArgs[cloudProviderKey] = cloudProviderValue
				}
				if _, ok := c.ControllerManager.ExtraArgs[cloudProviderKey]; !ok {
					c.ControllerManager.ExtraArgs[cloudProviderKey] = cloudProviderValue
				}

				// ensure projected token endpoints are enabled by configuring
				// issuer and signing key
				serviceAccountIssuerKey := "service-account-issuer"
				if _, ok := c.APIServer.ExtraArgs[serviceAccountIssuerKey]; !ok {
					apiServerURL := url.URL{
						Scheme: "https",
						Host: fmt.Sprintf(
							"%s:%d",
							s.scope.Cluster.Spec.ControlPlaneEndpoint.Host,
							s.scope.Cluster.Spec.ControlPlaneEndpoint.Port,
						),
					}
					c.APIServer.ExtraArgs[serviceAccountIssuerKey] = apiServerURL.String()
				}
				serviceAccountSigningKeyFileKey := "service-account-signing-key-file"
				if _, ok := c.APIServer.ExtraArgs[serviceAccountSigningKeyFileKey]; !ok {
					c.APIServer.ExtraArgs[serviceAccountSigningKeyFileKey] = "/etc/kubernetes/pki/sa.key"
				}

				// configure APIserver serving certificate
				extraNames := []string{"127.0.0.1", "localhost"}
				for _, lb := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
					extraNames = append(extraNames, lb.IPv4)
					extraNames = append(extraNames, lb.IPv6)
				}
				for _, name := range extraNames {
					if !stringSliceContains(c.APIServer.CertSANs, name) {
						c.APIServer.CertSANs = append(
							c.APIServer.CertSANs,
							name,
						)
					}
				}
				fmt.Println(c.APIServer.CertSANs)
			} else {
				record.Warnf(
					s.scope.HcloudMachine,
					"UnexpectedUserData",
					"UserData for a control plane comes without a ClusterConfiguration",
				)
			}
		}
	}

	if j := kubeadmConfig.JoinConfiguration; j != nil {
		if j.NodeRegistration.KubeletExtraArgs == nil {
			j.NodeRegistration.KubeletExtraArgs = make(map[string]string)
		}
		if _, ok := j.NodeRegistration.KubeletExtraArgs[cloudProviderKey]; !ok {
			j.NodeRegistration.KubeletExtraArgs[cloudProviderKey] = cloudProviderValue
		}
	}

	// TODO: Handle volumes

	if err := userData.SetKubeadmConfig(kubeadmConfig); err != nil {
		return nil, err
	}

	userDataBytes := bytes.NewBuffer(nil)
	if err := userData.WriteYAML(userDataBytes); err != nil {
		return nil, err
	}

	// update current server
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh server status")
	}

	var myTrue = true
	var myFalse = false

	name := s.scope.Name()
	opts := hcloud.ServerCreateOpts{
		Name:   name,
		Labels: s.createLabels(),
		Image: &hcloud.Image{
			ID: int(*s.scope.HcloudMachine.Status.ImageID),
		},
		Location: &hcloud.Location{
			Name: failureDomain,
		},
		ServerType: &hcloud.ServerType{
			Name: string(s.scope.HcloudMachine.Spec.Type),
		},
		Automount:        &myFalse,
		StartAfterCreate: &myTrue,
		UserData:         userDataBytes.String(),
		Volumes:          volumes,
	}

	// setup SSH keys
	sshKeySpecs := s.scope.HcloudMachine.Spec.SSHKeys
	if len(sshKeySpecs) == 0 {
		sshKeySpecs = s.scope.HcloudCluster.Spec.SSHKeys
	}
	sshKeys, _, err := s.scope.HcloudClient().ListSSHKeys(ctx, hcloud.SSHKeyListOpts{})
	if err != nil {
	}
	for _, sshKey := range sshKeys {
		var match bool
		for _, sshKeySpec := range sshKeySpecs {
			if sshKeySpec.Name != nil && *sshKeySpec.Name == sshKey.Name {
				match = true
			}
			if sshKeySpec.ID != nil && *sshKeySpec.ID == sshKey.ID {
				match = true
			}
		}
		if match {
			opts.SSHKeys = append(opts.SSHKeys, sshKey)
		}
	}

	// setup network if available
	if net := s.scope.HcloudCluster.Status.Network; net != nil {
		opts.Networks = []*hcloud.Network{{
			ID: net.ID,
		}}
	}

	var actualServer *hcloud.Server

	if len(actualServers) == 0 {
		fmt.Printf("Found no actualServer, server name is %s", name)
		if res, _, err := s.scope.HcloudClient().CreateServer(s.scope.Ctx, opts); err != nil {
			return nil, errors.Wrap(err, "failed to create server")
		} else {
			record.Eventf(
				s.scope.HcloudMachine,
				"SuccessfulCreate",
				"Created new server with id %d",
				res.Server.ID,
			)
			actualServer = res.Server
		}
	} else if len(actualServers) == 1 {
		actualServer = actualServers[0]
	} else {
		return nil, errors.New("found more than one actual servers")
	}

	if err := setStatusFromAPI(&s.scope.HcloudMachine.Status, actualServer); err != nil {
		return nil, errors.New("error setting status")
	}

	// wait for server being running
	if actualServer.Status != hcloud.ServerStatusRunning {
		s.scope.V(1).Info("server not in running state", "server", actualServer.Name, "status", actualServer.Status)
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	providerID := fmt.Sprintf("hcloud://%d", actualServer.ID)

	if !s.scope.IsControlPlane() {
		s.scope.HcloudMachine.Spec.ProviderID = &providerID
		s.scope.HcloudMachine.Status.Ready = true
		return nil, nil
	}
	// TODO: Check if we need this here, or if it is redundant as the server cannot be added anyway
	// add server as target to load balancer
	/*
		if err := s.addServerToLoadBalancer(actualServer); err != nil {
			return nil, errors.New("error adding server as target to load balancer")
		}
	*/

	// check if at least one of the adresses is ready
	var errors []error
	for _, address := range s.scope.HcloudMachine.Status.Addresses {
		if address.Type != corev1.NodeExternalIP && address.Type != corev1.NodeExternalDNS {
			continue
		}

		clientConfig, err := s.scope.ClientConfigWithAPIEndpoint(clusterv1.APIEndpoint{
			/*
				Host: address.Address,
				Port: s.scope.ControlPlaneAPIEndpointPort(),
			*/
		})
		if err != nil {
			return nil, err
		}

		if err := scope.IsControlPlaneReady(ctx, clientConfig); err != nil {
			errors = append(errors, err)
		}

		s.scope.HcloudMachine.Spec.ProviderID = &providerID
		s.scope.HcloudMachine.Status.Ready = true
		return nil, nil
	}

	if err := errorutil.NewAggregate(errors); err != nil {
		record.Warnf(
			s.scope.HcloudMachine,
			"APIServerNotReady",
			"Health check for API server failed: %s",
			err,
		)
	}
	return nil, fmt.Errorf("Not usable Address found")
}

func (s *Service) Delete(ctx context.Context) (_ *ctrl.Result, err error) {
	// update current servers
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh server status")
	}

	var actionWait []*hcloud.Server
	var actionShutdown []*hcloud.Server
	var actionDelete []*hcloud.Server

	for _, server := range actualServers {
		switch status := server.Status; status {
		case hcloud.ServerStatusRunning:
			actionShutdown = append(actionShutdown, server)
		case hcloud.ServerStatusOff:
			actionDelete = append(actionDelete, server)
		default:
			actionWait = append(actionWait, server)
		}
	}

	// shutdown servers
	for _, server := range actionShutdown {
		if _, _, err := s.scope.HcloudClient().ShutdownServer(ctx, server); err != nil {
			return nil, errors.Wrap(err, "failed to shutdown server")
		}
		actionWait = append(actionWait, server)
	}

	// delete servers that need delete
	for _, server := range actionDelete {
		if _, err := s.scope.HcloudClient().DeleteServer(ctx, server); err != nil {
			return nil, errors.Wrap(err, "failed to delete server")
		}
	}

	var result *ctrl.Result
	if len(actionWait) > 0 {
		result = &ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}
	}

	return result, nil
}

func setStatusFromAPI(status *infrav1.HcloudMachineStatus, server *hcloud.Server) error {
	status.ServerState = infrav1.HcloudServerState(server.Status)
	status.Addresses = []corev1.NodeAddress{}

	if ip := server.PublicNet.IPv4.IP.String(); ip != "" {
		status.Addresses = append(
			status.Addresses,
			corev1.NodeAddress{
				Type:    corev1.NodeExternalIP,
				Address: ip,
			},
		)
	}

	if ip := server.PublicNet.IPv6.IP; ip.IsGlobalUnicast() {
		ip[15] += 1
		status.Addresses = append(
			status.Addresses,
			corev1.NodeAddress{
				Type:    corev1.NodeExternalIP,
				Address: ip.String(),
			},
		)
	}

	for _, net := range server.PrivateNet {
		status.Addresses = append(
			status.Addresses,
			corev1.NodeAddress{
				Type:    corev1.NodeInternalIP,
				Address: net.IP.String(),
			},
		)

	}

	return nil
}

// actualStatus gathers all matching server instances, matched by tag
func (s *Service) actualStatus(ctx context.Context) ([]*hcloud.Server, error) {
	opts := hcloud.ServerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.createLabels())
	servers, err := s.scope.HcloudClient().ListServers(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	return servers, nil

}
