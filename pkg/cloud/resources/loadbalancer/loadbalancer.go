package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/storage/names"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/utils"
)

type Service struct {
	scope *scope.ClusterScope
}

func NewService(scope *scope.ClusterScope) *Service {
	return &Service{
		scope: scope,
	}
}

var errNotImplemented = errors.New("Not implemented")

type intSlice []int

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
	if len(lb.PrivateNet) > 0 {
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

	hasNetwork := len(lb.PrivateNet) > 0

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
		HasNetwork: hasNetwork,
	}
	return status, nil
}

func (s *Service) Reconcile(ctx context.Context) (err error) {

	s.scope.V(3).Info("Reconcile load balancer")

	// update current status
	loadBalancerStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancer")
	}
	if loadBalancerStatus != nil {
		s.scope.HcloudCluster.Status.ControlPlaneLoadBalancer = *loadBalancerStatus
		if loadBalancerStatus.HasNetwork == false {
			s.attachLoadBalancerToNetwork(ctx)
		}
		// TODO: Check if targets are up-to-date and add/delete some if needed
	}

	if loadBalancerStatus == nil {
		if _, err := s.createLoadBalancer(ctx, s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancer); err != nil {
			return errors.Wrap(err, "failed to create load balancer")
		}
	}

	// update current status
	loadBalancerStatus, err = s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancers")
	}
	s.scope.HcloudCluster.Status.ControlPlaneLoadBalancer = *loadBalancerStatus

	return nil
}

func (s *Service) createLoadBalancer(ctx context.Context, spec infrav1.HcloudLoadBalancerSpec) (*infrav1.HcloudLoadBalancerStatus, error) {

	s.scope.V(2).Info("Create a new loadbalancer", "algorithm type", spec.Algorithm)

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

	if hc.Status.Network == nil {
		return nil, errors.New("no network set on the cluster")
	}
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

	var mybool = false
	var defaultPort = int(s.scope.ControlPlaneAPIEndpointPort())
	var listenPort = defaultPort
	// If controlPlaneEndpoint has been specified (or set) then use it
	if s.scope.HcloudCluster.Spec.ControlPlaneEndpoint != nil {
		listenPort = int(s.scope.HcloudCluster.Spec.ControlPlaneEndpoint.Port)
	}

	kubeapiservice := hcloud.LoadBalancerCreateOptsService{
		Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
		ListenPort:      &listenPort,
		DestinationPort: &defaultPort,
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
		Services:         []hcloud.LoadBalancerCreateOptsService{kubeapiservice},
	}

	lb, _, err := s.scope.HcloudClient().CreateLoadBalancer(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating load balancer: %s", err)
	}

	status, err := s.apiToStatus(lb.LoadBalancer)
	if err != nil {
		return nil, fmt.Errorf("error converting load balancer: %s", err)
	}

	return status, nil
}

func (s *Service) attachLoadBalancerToNetwork(ctx context.Context) error {
	lb, err := GetLoadBalancer(s.scope)
	if err != nil {
		errors.Wrap(err, "failed to get load balancer")
	}

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

	return errors.Wrap(err, "failed to attach load balancer to network")
}

func (s *Service) deleteLoadBalancer(ctx context.Context, status infrav1.HcloudLoadBalancerStatus) error {

	// ensure deleted load balancer is actually owned by us
	clusterTagKey := infrav1.ClusterTagKey(s.scope.HcloudCluster.Name)
	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}

	opts := hcloud.LoadBalancerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(labels)

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "Failed to list load balancers")
	}

	var lb *hcloud.LoadBalancer
	for _, loadBalancer := range loadBalancers {
		if loadBalancer.ID == status.ID {
			lb = loadBalancer
		}
	}
	if lb == nil {
		return fmt.Errorf("No load balancer found with ID %v", status.ID)
	}

	if lb.Labels == nil || infrav1.ResourceLifecycle(lb.Labels[clusterTagKey]) != infrav1.ResourceLifecycleOwned {
		s.scope.V(3).Info("Ignore request to delete load balancer, as it is not owned", "id", lb.ID, "name", lb.Name)
		return nil
	}

	_, err = s.scope.HcloudClient().DeleteLoadBalancer(ctx, lb)

	s.scope.V(2).Info("Delete load balancer", "id", lb.ID, "name", lb.Name)

	return err
}

// Delete calls the deleteLoadBalancer function to delete one load balancer after another
func (s *Service) Delete(ctx context.Context) (err error) {
	// update current status
	loadBalancerStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancer")
	}
	if loadBalancerStatus == nil {
		return nil
	}

	if err := s.deleteLoadBalancer(ctx, *loadBalancerStatus); err != nil {
		return errors.Wrap(err, "failed to delete load balancer")
	}

	return nil
}

func GetLoadBalancer(s *scope.ClusterScope) (*hcloud.LoadBalancer, error) {

	clusterTagKey := infrav1.ClusterTagKey(s.HcloudCluster.Name)
	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}

	opts := hcloud.LoadBalancerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(labels)
	loadBalancers, err := s.HcloudClient().ListLoadBalancers(s.Ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list load balancers")
	}

	if len(loadBalancers) == 0 {
		return nil, fmt.Errorf("No main load balancer exists")
	} else if len(loadBalancers) > 1 {
		return nil, fmt.Errorf("Too many, i.e. %v, load balancers exist", len(loadBalancers))
	} else {
		lb := loadBalancers[0]
		return lb, nil
	}
}

// actualStatus gathers the load balancer with
// appropriate tag and converts it into the status object
func (s *Service) actualStatus(ctx context.Context) (*infrav1.HcloudLoadBalancerStatus, error) {

	clusterTagKey := infrav1.ClusterTagKey(s.scope.HcloudCluster.Name)
	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}
	opts := hcloud.LoadBalancerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(labels)

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list load balancers")
	}

	if len(loadBalancers) > 1 {
		return nil, fmt.Errorf("Found %v loadbalancers in Hcloud", len(loadBalancers))
	} else if len(loadBalancers) == 0 {
		return nil, nil
	}

	lbStatus, err := s.apiToStatus(loadBalancers[0])
	if err != nil {
		errors.Wrap(err, "failed to convert load balancer to status object")
	}

	return lbStatus, nil

}
