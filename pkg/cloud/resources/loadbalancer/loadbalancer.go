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

func apiToStatus(lb *hcloud.LoadBalancer) (*infrav1.HcloudLoadBalancerStatus, error) {

	ipv4 := lb.PublicNet.IPv4.IP.String()
	ipv6 := lb.PublicNet.IPv6.IP.String()
	network := fmt.Sprintf("%s/%s", lb.PrivateNet[0].Network.IPRange.IP, lb.PrivateNet[0].Network.IPRange.Mask)

	var algType infrav1.HcloudLoadBalancerAlgorithmType

	if lb.Algorithm.Type == hcloud.LoadBalancerAlgorithmTypeRoundRobin {
		algType = infrav1.HcloudLoadBalancerAlgorithmTypeRoundRobin
	} else if lb.Algorithm.Type == hcloud.LoadBalancerAlgorithmTypeLeastConnections {
		algType = infrav1.HcloudLoadBalancerAlgorithmTypeLeastConnections
	} else {
		return nil, fmt.Errorf("Unknown load balancer algorithm type: %s", lb.Algorithm.Type)
	}

	status := &infrav1.HcloudLoadBalancerStatus{
		ID:        lb.ID,
		Name:      lb.Name,
		Type:      lb.LoadBalancerType.Name,
		IPv4:      ipv4,
		IPv6:      ipv6,
		Labels:    lb.Labels,
		Algorithm: algType,
		Network:   network,
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

	// gather ip type
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
	clusterTagKey := infrav1.ClusterTagKey(hc.Name)

	if len(hc.Status.Locations) == 0 {
		return nil, errors.New("no locations set on the cluster")
	}
	location := &hcloud.Location{Name: string(hc.Status.Locations[0])}
	name := names.SimpleNameGenerator.GenerateName(hc.Name + "-loadbalancer-")

	// defaults to the smalles load balancer, 25 targets should be enough for the control-planes
	hcloudToken := s.scope.HcloudClient().Token()
	hclient := hcloud.NewClient(hcloud.WithToken(hcloudToken))
	loadBalancerType, _, err := hclient.LoadBalancerType.GetByName(context.Background(), spec.Type)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find load balancer type")
	}

	networkID := s.scope.HcloudCluster.Status.Network.ID
	network, _, err := hclient.Network.GetByID(context.Background(), networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to find network with ID %s", networkID)
	}

	opts := hcloud.LoadBalancerCreateOpts{
		LoadBalancerType: loadBalancerType,
		Name:             name,
		Algorithm:        loadBalancerAlgorithm,
		Location:         location,
		Network:          network,
		Labels: map[string]string{
			clusterTagKey: string(infrav1.ResourceLifecycleOwned),
		},
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
	if status.Labels == nil || infrav1.ResourceLifecycle(status.Labels[clusterTagKey]) != infrav1.ResourceLifecycleOwned {
		s.scope.V(3).Info("Ignore request to delete load balancer, as it is not owned", "id", status.ID, "name", status.Name)
		return nil
	}
	_, err := s.scope.HcloudClient().DeleteLoadBalancer(ctx, &hcloud.LoadBalancer{ID: status.ID})
	s.scope.V(2).Info("Delete load balancer", "id", status.ID, "name", status.Name)
	return err
}

func (s *Service) Delete(ctx context.Context) (err error) {
	// update current status
	loadBalancerStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh load balancer")
	}

	for _, status := range loadBalancerStatus {
		if err := s.deleteLoadBalancer(ctx, status); err != nil {
			return errors.Wrap(err, "failed to delete load balancer")
		}
	}

	return nil
}

func (s *Service) compare(actualStatus []infrav1.HcloudLoadBalancerStatus) (needCreation []infrav1.HcloudLoadBalancerSpec, needDeletion []infrav1.HcloudLoadBalancerStatus) {
	var matchedSpecToStatusMap = make(map[int]*infrav1.HcloudLoadBalancerStatus)
	var matchedStatusToSpecMap = make(map[int]*infrav1.HcloudLoadBalancerSpec)

	for posSpec, ipSpec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancers {
		for posStatus, ipStatus := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
			if _, ok := matchedStatusToSpecMap[posStatus]; ok {
				continue
			}
			matchedStatusToSpecMap[posStatus] = &ipSpec
			matchedSpecToStatusMap[posSpec] = &ipStatus
		}
	}

	// load balancers to create
	for pos, spec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancers {
		if _, ok := matchedSpecToStatusMap[pos]; ok {
			continue
		}
		needCreation = append(needCreation, spec)
	}

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
	ipStatusByID := make(map[int]*infrav1.HcloudLoadBalancerStatus)
	for pos := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
		ipStatus := &s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers[pos]
		ipStatusByID[ipStatus.ID] = ipStatus
		ids = append(ids, ipStatus.ID)
	}
	for _, ipSpec := range s.scope.HcloudCluster.Spec.ControlPlaneLoadBalancers {
		if ipSpec.ID != nil {
			ids = append(ids, *ipSpec.ID)
		}
	}

	// refresh existing load balancers
	clusterTagKey := infrav1.ClusterTagKey(s.scope.HcloudCluster.Name)
	ipStatuses, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing load balancers: %w", err)
	}
	for _, ipStatus := range ipStatuses {
		_, ok := ipStatus.Labels[clusterTagKey]
		if !ok && !ids.contains(ipStatus.ID) {
			continue
		}

		apiStatus, err := apiToStatus(ipStatus)
		if err != nil {
			return nil, fmt.Errorf("error converting LoadBalancer API to status: %w", err)
		}
		ipStatusByID[apiStatus.ID] = apiStatus
	}

	ids = []int{}
	for id := range ipStatusByID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var loadBalancers []infrav1.HcloudLoadBalancerStatus

	for _, id := range ids {
		status := ipStatusByID[id]
		loadBalancers = append(
			loadBalancers,
			*status,
		)
	}

	return loadBalancers, nil
}
