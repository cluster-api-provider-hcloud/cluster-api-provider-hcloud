package loadbalancer

import (
	"context"
	"fmt"
	"sort"

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
func apiToStatus(lb *hcloud.LoadBalancer) (*infrav1.HcloudLoadBalancerStatus, error) {

	ipv4 := lb.PublicNet.IPv4.IP.String()
	ipv6 := lb.PublicNet.IPv6.IP.String()

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
		ID:        lb.ID,
		Name:      lb.Name,
		Type:      lb.LoadBalancerType.Name,
		IPv4:      ipv4,
		IPv6:      ipv6,
		Labels:    lb.Labels,
		Algorithm: algType,
		Targets:   targetIDs,
	}
	return status, nil
}

func (s *Service) Reconcile(ctx context.Context) (err error) {

	// update current status
	loadBalancerStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancers")
	}
	s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers = loadBalancerStatus

	s.scope.V(3).Info("Reconcile load balancers")

	// Compares spec and actual status to find out if load balancers need to be created/deleted
	needCreation, needDeletion := s.compare(loadBalancerStatus)

	if len(needCreation) == 0 && len(needDeletion) == 0 {
		return nil
	}

	for _, spec := range needCreation {
		if _, err := s.createLoadBalancer(ctx, spec); err != nil {
			return errors.Wrap(err, "failed to create load balancer")
		}
	}

	for _, status := range needDeletion {
		if err := s.deleteLoadBalancer(ctx, status); err != nil {
			return errors.Wrap(err, "failed to delete load balancer")
		}
	}

	// update current status
	loadBalancerStatus, err = s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancers")
	}
	s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers = loadBalancerStatus

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

	name := names.SimpleNameGenerator.GenerateName(hc.Name + "-loadbalancer-")

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
	var myDestPort = int(s.scope.ControlPlaneAPIEndpointPort())
	var myListenPort *int

	// If listen port is specified in the specs we use that, otherwise we choose the destination port
	// as listen port
	if spec.ListenPort != nil {
		myListenPort = spec.ListenPort
	} else {
		myListenPort = &myDestPort
	}

	kubeapiservice := hcloud.LoadBalancerCreateOptsService{
		Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
		ListenPort:      myListenPort,
		DestinationPort: &myDestPort,
		Proxyprotocol:   &mybool,
	}

	clusterTagKey := infrav1.ClusterTagKey(hc.Name)

	// The first load balancer automatically becomes the main one that has the control
	// planes as targets.
	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}
	if len(s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers) == 0 {
		labels["type"] = "main"
	} else {
		labels["type"] = "other"
	}

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

	status, err := apiToStatus(lb.LoadBalancer)
	if err != nil {
		return nil, fmt.Errorf("error converting load balancer: %s", err)
	}

	return status, nil
}

func (s *Service) deleteLoadBalancer(ctx context.Context, status infrav1.HcloudLoadBalancerStatus) error {

	// ensure deleted load balancer is actually owned by us
	clusterTagKey := infrav1.ClusterTagKey(s.scope.HcloudCluster.Name)

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
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

	for _, lbStatus := range loadBalancerStatus {
		if err := s.deleteLoadBalancer(ctx, lbStatus); err != nil {
			return errors.Wrap(err, "failed to delete load balancer")
		}
	}

	return nil
}

func GetMainLoadBalancer(s *scope.ClusterScope, ctx context.Context) (*hcloud.LoadBalancer, error) {
	labels := map[string]string{
		"type": "main",
	}

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

func (s *Service) compare(actualStatus []infrav1.HcloudLoadBalancerStatus) (needCreation []infrav1.HcloudLoadBalancerSpec, needDeletion []infrav1.HcloudLoadBalancerStatus) {
	var matchedSpecToStatusMap = make(map[int]*infrav1.HcloudLoadBalancerStatus)
	var matchedStatusToSpecMap = make(map[int]*infrav1.HcloudLoadBalancerSpec)

	for posSpec, lbSpec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancers {
		for posStatus, lbStatus := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
			if _, ok := matchedStatusToSpecMap[posStatus]; ok {
				continue
			}
			matchedStatusToSpecMap[posStatus] = &lbSpec
			matchedSpecToStatusMap[posSpec] = &lbStatus
		}
	}

	// load balancers to create
	for pos, spec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancers {
		if _, ok := matchedSpecToStatusMap[pos]; ok {
			continue
		}
		needCreation = append(needCreation, spec)
	}

	// load balancers to delete
	for pos, status := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
		if _, ok := matchedStatusToSpecMap[pos]; ok {
			continue
		}
		needDeletion = append(needDeletion, status)
	}

	return needCreation, needDeletion
}

// actualStatus gathers all load balancers referenced by ID on object or a
// appropriate tag and converts them into the status object
func (s *Service) actualStatus(ctx context.Context) ([]infrav1.HcloudLoadBalancerStatus, error) {
	// index existing status entries
	var ids intSlice
	lbStatusByID := make(map[int]*infrav1.HcloudLoadBalancerStatus)
	for pos := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
		lbStatus := &s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers[pos]
		lbStatusByID[lbStatus.ID] = lbStatus
		ids = append(ids, lbStatus.ID)
	}
	for _, lbSpec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancers {
		if lbSpec.ID != nil {
			ids = append(ids, *lbSpec.ID)
		}
	}

	// refresh existing load balancers
	clusterTagKey := infrav1.ClusterTagKey(s.scope.HcloudCluster.Name)
	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing load balancers: %w", err)
	}
	for _, lb := range loadBalancers {
		_, ok := lb.Labels[clusterTagKey]
		if !ok && !ids.contains(lb.ID) {
			continue
		}

		lbStatus, err := apiToStatus(lb)
		if err != nil {
			return nil, fmt.Errorf("error converting LoadBalancer API to status: %w", err)
		}
		lbStatusByID[lbStatus.ID] = lbStatus
	}

	ids = []int{}
	for id := range lbStatusByID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var lbStatuses []infrav1.HcloudLoadBalancerStatus

	for _, id := range ids {
		status := lbStatusByID[id]
		lbStatuses = append(
			lbStatuses,
			*status,
		)
	}

	return lbStatuses, nil
}
