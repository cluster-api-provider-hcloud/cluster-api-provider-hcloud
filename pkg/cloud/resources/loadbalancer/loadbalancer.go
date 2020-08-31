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
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/record"
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

	/*
		var network string
		if len(lb.PrivateNet) > 0 {
			fmt.Printf(lb.PrivateNet[0].Network.IPRange.IP.String())
			network = fmt.Sprintf("%s/%s", lb.PrivateNet[0].Network.IPRange.IP.String(), string(lb.PrivateNet[0].Network.IPRange.Mask))
		}
	*/
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
		//Network:   network,
		Targets: targetIDs,
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

	// update targets of load balancer
	needCreationTargets, needDeletionTargets, err := s.compareServerTargets()
	if err != nil {
		return errors.Wrap(err, "failed to update server targets")
	}

	if len(needCreationTargets) == 0 && len(needDeletionTargets) == 0 {
		return nil
	}

	for _, server := range needCreationTargets {
		if err := s.addServerToLoadBalancer(ctx, server); err != nil {
			return errors.Wrap(err, "failed to add server to load balancer")
		}
	}

	for _, server := range needDeletionTargets {
		if err := s.deleteServerOfLoadBalancer(ctx, server); err != nil {
			return errors.Wrap(err, "failed to delete server of load balancer")
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
	clusterTagKey := infrav1.ClusterTagKey(hc.Name)
	// defaults to the smalles load balancer, 25 targets should be enough for the control-planes

	loadBalancerType, _, err := s.scope.HcloudClient().GetLoadBalancerTypeByName(ctx, spec.Type)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find load balancer type")
	}

	if len(hc.Status.Locations) == 0 {
		return nil, errors.New("no locations set on the cluster")
	}
	location := &hcloud.Location{Name: string(hc.Status.Locations[0])}

	var network *hcloud.Network
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
	var myInt = int(s.scope.ControlPlaneAPIEndpointPort())

	kubeapiservice := hcloud.LoadBalancerCreateOptsService{
		Protocol:        hcloud.LoadBalancerServiceProtocolTCP,
		ListenPort:      &myInt,
		DestinationPort: &myInt,
		Proxyprotocol:   &mybool,
	}

	labels := map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)}
	if len(s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers) == 0 {
		labels["type"] = "main"
	} else {
		labels["type"] = "other"
	}

	//loadBalancerCreateOptsTarget := hcloud.LoadBalancerCreateOptsTarget{Type: hcloud.LoadBalancerTargetTypeServer}

	opts := hcloud.LoadBalancerCreateOpts{
		LoadBalancerType: loadBalancerType,
		Name:             name,
		Algorithm:        loadBalancerAlgorithm,
		Location:         location,
		Network:          network,
		Labels:           labels,
		Services:         []hcloud.LoadBalancerCreateOptsService{kubeapiservice},
		//Targets:          []hcloud.LoadBalancerCreateOptsTarget{loadBalancerCreateOptsTarget},
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

// deleting loadbalancer
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

func (s *Service) compareServerTargets() (needCreation []*hcloud.Server, needDeletion []*hcloud.Server, err error) {
	controlPlanes, err := s.listControlPlanes()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list all control planes")
	}

	hcloudToken := s.scope.HcloudClient().Token()
	hclient := hcloud.NewClient(hcloud.WithToken(hcloudToken))

	var controlPlaneStatusIDs intSlice
	controlPlaneStatusIDs = s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers[0].Targets

	var controlPlaneIDs intSlice

	for _, cp := range controlPlanes {

		controlPlaneIDs = append(controlPlaneIDs, cp.ID)
		// Check whether control plane is in target set of loadbalancer
		// If not than add it

		if !controlPlaneStatusIDs.contains(cp.ID) {
			needCreation = append(needCreation, cp)
		}
	}

	for _, id := range controlPlaneStatusIDs {
		if !controlPlaneIDs.contains(id) {
			server, _, err := hclient.Server.GetByID(context.Background(), id)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to get server")
			}
			needDeletion = append(needDeletion, server)
		}
	}
	return needCreation, needDeletion, nil
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

	for pos, status := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
		if _, ok := matchedStatusToSpecMap[pos]; ok {
			continue
		}
		needDeletion = append(needDeletion, status)
	}

	return needCreation, needDeletion
}

func (s *Service) labels() map[string]string {
	return map[string]string{
		infrav1.ClusterTagKey(s.scope.HcloudCluster.Name): string(infrav1.ResourceLifecycleOwned),
		"machine_type": "control_plane",
	}
}

func (s *Service) listControlPlanes() ([]*hcloud.Server, error) {
	opts := hcloud.ServerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.labels())
	servers, err := s.scope.HcloudClient().ListServers(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	return servers, nil

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
	lbStatuses, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing load balancers: %w", err)
	}
	for _, lbStatus := range lbStatuses {
		_, ok := lbStatus.Labels[clusterTagKey]
		if !ok && !ids.contains(lbStatus.ID) {
			continue
		}

		apiStatus, err := apiToStatus(lbStatus)
		if err != nil {
			return nil, fmt.Errorf("error converting LoadBalancer API to status: %w", err)
		}
		lbStatusByID[apiStatus.ID] = apiStatus
	}

	ids = []int{}
	for id := range lbStatusByID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var loadBalancers []infrav1.HcloudLoadBalancerStatus

	for _, id := range ids {
		status := lbStatusByID[id]
		loadBalancers = append(
			loadBalancers,
			*status,
		)
	}

	return loadBalancers, nil
}

func (s *Service) addServerToLoadBalancer(ctx context.Context, server *hcloud.Server) error {

	myBool := true
	loadBalancerAddServerTargetOpts := hcloud.LoadBalancerAddServerTargetOpts{Server: server, UsePrivateIP: &myBool}

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
	if err != nil {
		return err
	}
	// This only works if there is only one load balancer
	lb := loadBalancers[0]
	_, _, err = s.scope.HcloudClient().AddTargetServerToLoadBalancer(ctx, loadBalancerAddServerTargetOpts, lb)
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

func (s *Service) deleteServerOfLoadBalancer(ctx context.Context, server *hcloud.Server) error {

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
	if err != nil {
		return err
	}
	// This only works if there is only one load balancer
	lb := loadBalancers[0]
	_, _, err = s.scope.HcloudClient().DeleteTargetServerOfLoadBalancer(ctx, lb, server)
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
