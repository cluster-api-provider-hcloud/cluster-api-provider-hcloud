package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/server/userdata"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/record"
)

type Service struct {
	scope *scope.BareMetalMachineScope
}

func NewService(scope *scope.BareMetalMachineScope) *Service {
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

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {

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

	if j := kubeadmConfig.JoinConfiguration; j != nil {
		if j.NodeRegistration.KubeletExtraArgs == nil {
			j.NodeRegistration.KubeletExtraArgs = make(map[string]string)
		}
		if _, ok := j.NodeRegistration.KubeletExtraArgs[cloudProviderKey]; !ok {
			j.NodeRegistration.KubeletExtraArgs[cloudProviderKey] = cloudProviderValue
		}
	}

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

	// check if at least one of the adresses is ready
	var errors []error
	for _, address := range s.scope.HcloudMachine.Status.Addresses {
		if address.Type != corev1.NodeExternalIP && address.Type != corev1.NodeExternalDNS {
			continue
		}

		clientConfig, err := s.scope.ClientConfigWithAPIEndpoint(clusterv1.APIEndpoint{
			Host: address.Address,
			Port: s.scope.ControlPlaneAPIEndpointPort(),
		})
		if err != nil {
			return nil, err
		}

		// Need to have a private network if added to load balancer
		if len(actualServer.PrivateNet) == 0 {
			return &ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

		if err := s.addServerToLoadBalancer(ctx, actualServer); err != nil {
			errors = append(errors, err)
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
		if server.Labels["machine_type"] == "control_plane" {
			s.deleteServerOfLoadBalancer(ctx, server)
		}

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

type BareMetalServer struct {
	ServerIP   string   `json:"server_ip"`
	ServerID   int      `json:"server_number"`
	ServerName string   `json:"server_name"`
	Product    string   `json:"product"`
	DataCenter string   `json:"dc"`
	Traffic    string   `json:"traffic"`
	Status     string   `json:"status"`
	Cancelled  bool     `json:"cancelled"`
	PaidUntil  string   `json:"paid_until"`
	IPs        []string `json:"ip"`
	Subnets    []string `json:"subnet"`
}

// actualStatus gathers all matching server instances, matched by tag
func (s *Service) actualStatus(ctx context.Context) ([]BareMetalServer, error) {

	resp, err := http.Get("https://robot-ws.your-server.de/server")
	if err != nil {
		return nil, err
	}

	var serverList []BareMetalServer

	json.NewDecoder(resp.Body).Decode(&serverList)
	if err != nil {
		return nil, err
	}

	var actualServers []BareMetalServer

	for _, server := range serverList {
		splitName := strings.Split(server.ServerName, "_")
		if splitName[0] == s.scope.Cluster.Name && splitName[1] == s.scope.BareMetalMachine.Name {
			actualServers = append(actualServers, server)
		}
	}

	return actualServers, nil
}
