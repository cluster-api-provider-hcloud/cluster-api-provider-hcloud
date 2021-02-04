package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/storage/names"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/utils"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/scope"
)

type Service struct {
	scope *scope.ClusterScope
}

func NewService(scope *scope.ClusterScope) *Service {
	return &Service{
		scope: scope,
	}
}

type intSlice []int

func (s *Service) Reconcile(ctx context.Context) (err error) {

	s.scope.V(3).Info("Reconcile load balancer")
	var lbStatus *infrav1.HcloudLoadBalancerStatus
	// find load balancer
	lb, err := FindLoadBalancer(s.scope)
	if err != nil {
		return errors.Wrap(err, "failed to find load balancer")
	}

	if lb != nil {
		lbStatus, err = s.apiToStatus(lb)
		if err != nil {
			return errors.Wrap(err, "failed to obtain load balancer status")
		}
		s.scope.HcloudCluster.Status.ControlPlaneLoadBalancer = *lbStatus
		if s.scope.HcloudCluster.Status.Network != nil && len(lb.PrivateNet) == 0 {
			if err := s.attachLoadBalancerToNetwork(ctx, lb); err != nil {
				return errors.Wrap(err, "failed to attach load balancer to network")
			}
		}
		// TODO: Check if targets are up-to-date (if machine reconciler is not triggered anyway)
	}

	if lb == nil {
		if lb, err = s.createLoadBalancer(ctx, s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer); err != nil {
			return errors.Wrap(err, "failed to create load balancer")
		}
	}

	// update current status
	lbStatus, err = s.apiToStatus(lb)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancer status")
	}
	s.scope.HcloudCluster.Status.ControlPlaneLoadBalancer = *lbStatus

	return nil
}

func (s *Service) createLoadBalancer(ctx context.Context, spec infrav1.HcloudLoadBalancerSpec) (*hcloud.LoadBalancer, error) {

	s.scope.V(1).Info("Create a new loadbalancer", "algorithm type", spec.Algorithm)

	// gather algorithm type
	var algType hcloud.LoadBalancerAlgorithmType
	if spec.Algorithm == infrav1.HcloudLoadBalancerAlgorithmTypeRoundRobin {
		algType = hcloud.LoadBalancerAlgorithmTypeRoundRobin
	} else if spec.Algorithm == infrav1.HcloudLoadBalancerAlgorithmTypeLeastConnections {
		algType = hcloud.LoadBalancerAlgorithmTypeLeastConnections
	} else {
		return nil, fmt.Errorf("error invalid load balancer algorithm type: %s", spec.Algorithm)
	}

	loadBalancerAlgorithm := (&hcloud.LoadBalancerAlgorithm{Type: algType})

	hc := s.scope.HcloudCluster

	name := names.SimpleNameGenerator.GenerateName(hc.Name + "-kube-apiserver-")
	if s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer.Name != nil {
		name = *s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer.Name
	}

	// Get the Hetzner cloud object of load balancer type
	loadBalancerType, _, err := s.scope.HcloudClient().GetLoadBalancerTypeByName(ctx, spec.Type)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find load balancer type")
	}

	if len(hc.Status.Locations) == 0 {
		return nil, errors.New("no locations set on the cluster")
	}
	location := &hcloud.Location{Name: string(hc.Status.Locations[0])}

	var network *hcloud.Network

	if hc.Status.Network != nil {
		networkID := s.scope.HcloudCluster.Status.Network.ID
		networks, err := s.scope.HcloudClient().ListNetworks(ctx, hcloud.NetworkListOpts{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list networks")
		}
		var i = false
		for _, net := range networks {
			if net.ID == networkID {
				network = net
				i = true
			}
		}
		if i == false {
			return nil, fmt.Errorf("could not find network with ID %v", networkID)
		}
	}

	var mybool bool
	// The first service in the list is the one of kubeAPI
	kubeAPISpec := s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer.Services[0]

	createServiceOpts := hcloud.LoadBalancerCreateOptsService{
		Protocol:        hcloud.LoadBalancerServiceProtocol(kubeAPISpec.Protocol),
		ListenPort:      &kubeAPISpec.ListenPort,
		DestinationPort: &kubeAPISpec.DestinationPort,
		Proxyprotocol:   &mybool,
	}

	clusterTagKey := infrav1.ClusterTagKey(hc.Name)

	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}

	opts := hcloud.LoadBalancerCreateOpts{
		LoadBalancerType: loadBalancerType,
		Name:             name,
		Algorithm:        loadBalancerAlgorithm,
		Location:         location,
		Network:          network,
		Labels:           labels,
		Services:         []hcloud.LoadBalancerCreateOptsService{createServiceOpts},
	}

	res, _, err := s.scope.HcloudClient().CreateLoadBalancer(ctx, opts)
	if err != nil {
		s.scope.Recorder.Eventf(
			s.scope.HcloudCluster,
			corev1.EventTypeWarning,
			"FailedCreateLoadBalancer",
			"Failed to create load balancer: %s",
			err)
		return nil, fmt.Errorf("error creating load balancer: %s", err)
	}

	// If there is more than one service in the specs, add them here one after another
	// Adding all at the same time on creation led to an error that the source port is already in use
	if len(s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer.Services) > 1 {
		for _, spec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer.Services[1:] {
			serviceOpts := hcloud.LoadBalancerAddServiceOpts{
				Protocol:        hcloud.LoadBalancerServiceProtocol(spec.Protocol),
				ListenPort:      &spec.ListenPort,
				DestinationPort: &spec.DestinationPort,
				Proxyprotocol:   &mybool,
			}
			_, _, err := s.scope.HcloudClient().AddServiceToLoadBalancer(ctx, res.LoadBalancer, serviceOpts)
			if err != nil {
				return nil, fmt.Errorf("Error adding service to load balancer: %s", err)
			}
		}
	}
	s.scope.Recorder.Eventf(s.scope.HcloudCluster, "CreateLoadBalancer", corev1.EventTypeNormal, "Created load balancer")
	return res.LoadBalancer, nil
}

func (s *Service) attachLoadBalancerToNetwork(ctx context.Context, lb *hcloud.LoadBalancer) error {

	var network *hcloud.Network

	if s.scope.HcloudCluster.Status.Network == nil {
		return errors.New("no network set on the cluster")
	}
	networkID := s.scope.HcloudCluster.Status.Network.ID
	networks, err := s.scope.HcloudClient().ListNetworks(ctx, hcloud.NetworkListOpts{})
	if err != nil {
		return errors.Wrap(err, "failed to list networks")
	}
	var i = false
	for _, net := range networks {
		if net.ID == networkID {
			network = net
			i = true
		}
	}
	if i == false {
		return fmt.Errorf("could not find network with ID %v", networkID)
	}

	opts := hcloud.LoadBalancerAttachToNetworkOpts{
		Network: network,
	}

	_, _, err = s.scope.HcloudClient().AttachLoadBalancerToNetwork(ctx, lb, opts)
	if err != nil {
		s.scope.Recorder.Eventf(
			s.scope.HcloudCluster,
			"FailedAttachLoadBalancer",
			corev1.EventTypeWarning,
			"Failed to attach load balancer to network: %s",
			err)
		return errors.Wrap(err, "failed to attach load balancer to network")
	}
	return nil
}

func (s *Service) Delete(ctx context.Context) (err error) {
	// update current status
	lb, err := FindLoadBalancer(s.scope)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancer")
	}
	if lb == nil {
		s.scope.Recorder.Eventf(s.scope.HcloudCluster, "UnknownLoadBalancer", corev1.EventTypeNormal, "Found no load balancer to delete")
		return nil
	}

	if _, err := s.scope.HcloudClient().DeleteLoadBalancer(ctx, lb); err != nil {
		s.scope.Recorder.Eventf(s.scope.HcloudCluster, "FailedLoadBalancerDelete", corev1.EventTypeNormal, "Failed to delete load balancer: %s", err)
		return errors.Wrap(err, "failed to delete load balancer")
	}

	s.scope.Recorder.Eventf(s.scope.HcloudCluster, "DeleteLoadBalancer", corev1.EventTypeNormal, "Deleted load balancer")

	return nil
}

func FindLoadBalancer(scope *scope.ClusterScope) (*hcloud.LoadBalancer, error) {

	clusterTagKey := infrav1.ClusterTagKey(scope.HcloudCluster.Name)
	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}
	opts := hcloud.LoadBalancerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(labels)

	loadBalancers, err := scope.HcloudClient().ListLoadBalancers(scope.Ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list load balancers")
	}

	if len(loadBalancers) > 1 {
		return nil, fmt.Errorf("Found %v loadbalancers in Hcloud", len(loadBalancers))
	} else if len(loadBalancers) == 0 {
		return nil, nil
	}

	return loadBalancers[0], nil

}

func (s intSlice) contains(e int) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

// gets the information of the Hetzner load balancer object and returns it in our status object
func (s *Service) apiToStatus(lb *hcloud.LoadBalancer) (*infrav1.HcloudLoadBalancerStatus, error) {

	ipv4 := lb.PublicNet.IPv4.IP.String()
	ipv6 := lb.PublicNet.IPv6.IP.String()

	var internalIP string
	if s.scope.HcloudCluster.Status.Network != nil && len(lb.PrivateNet) > 0 {
		internalIP = lb.PrivateNet[0].IP.String()
	}

	var algType infrav1.HcloudLoadBalancerAlgorithmType

	if lb.Algorithm.Type == hcloud.LoadBalancerAlgorithmTypeRoundRobin {
		algType = infrav1.HcloudLoadBalancerAlgorithmTypeRoundRobin
	} else if lb.Algorithm.Type == hcloud.LoadBalancerAlgorithmTypeLeastConnections {
		algType = infrav1.HcloudLoadBalancerAlgorithmTypeLeastConnections
	} else {
		return nil, fmt.Errorf("Unknown load balancer algorithm type: %s", lb.Algorithm.Type)
	}

	var targetIDs []int

	for _, server := range lb.Targets {
		targetIDs = append(targetIDs, server.Server.Server.ID)
	}

	status := &infrav1.HcloudLoadBalancerStatus{
		ID:         lb.ID,
		Name:       lb.Name,
		Type:       lb.LoadBalancerType.Name,
		IPv4:       ipv4,
		IPv6:       ipv6,
		InternalIP: internalIP,
		Labels:     lb.Labels,
		Algorithm:  algType,
		Targets:    targetIDs,
	}
	return status, nil
}
