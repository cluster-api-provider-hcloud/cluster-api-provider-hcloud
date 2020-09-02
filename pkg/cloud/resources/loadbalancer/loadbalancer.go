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

	// update targets of the first, i.e. the main, load balancer
	needCreationTargets, needDeletionTargets, err := s.compareServerTargets(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update server targets")
	}
	fmt.Println("Creation targets")
	for _, target := range needCreationTargets {
		fmt.Println("target id: ", target.ID)
		fmt.Println("target label: ", target.Labels)
	}
	for _, server := range needCreationTargets {
		if err := s.addServerToLoadBalancer(ctx, server); err != nil {
			return errors.Wrap(err, "failed to add servers to load balancer")
		}
	}
	fmt.Println("Deletion targets")
	for _, target := range needDeletionTargets {
		fmt.Println("target id: ", target.ID)
		fmt.Println("target label: ", target.Labels)
	}
	for _, server := range needDeletionTargets {
		if err := s.deleteServerOfLoadBalancer(ctx, server); err != nil {
			return errors.Wrap(err, "failed to delete servers of load balancer")
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

	if status.Labels == nil || infrav1.ResourceLifecycle(status.Labels[clusterTagKey]) != infrav1.ResourceLifecycleOwned {
		s.scope.V(3).Info("Ignore request to delete load balancer, as it is not owned", "id", status.ID, "name", status.Name)
		return nil
	}

	_, err := s.scope.HcloudClient().DeleteLoadBalancer(ctx, &hcloud.LoadBalancer{ID: status.ID})

	s.scope.V(2).Info("Delete load balancer", "id", status.ID, "name", status.Name)

	return err
}

// Delete calls the deleteLoadBalancer function to delete one load balancer after another
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

func (s *Service) GetMainLoadBalancer(ctx context.Context) (*hcloud.LoadBalancer, error) {
	labels := map[string]string{
		"type": "main",
	}
	fmt.Println("Labels of load balancer: ", labels)
	opts := hcloud.LoadBalancerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(labels)
	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(s.scope.Ctx, opts)
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

// compareServerTargets checks for the main load balancer whether all the control planes are targets, or whether
// there are targets which are not control planes anymore and have to be deleted
func (s *Service) compareServerTargets(ctx context.Context) (needCreation []*hcloud.Server, needDeletion []*hcloud.Server, err error) {
	controlPlanes, err := s.listControlPlanes()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list all control planes")
	}

	if len(s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers) == 0 {
		return nil, nil, nil
	}
	fmt.Println("Try to get main loadbalancer")
	lb, err := s.GetMainLoadBalancer(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "did not find main load balancer")
	}
	fmt.Println("Got it")

	var controlPlaneStatusIDs intSlice

	i := 0
	for _, lbStatus := range s.scope.HcloudCluster.Status.ControlPlaneLoadBalancers {
		fmt.Println("ID of load balancer: ", lb.ID)
		fmt.Println("ID of load balancer status: ", lbStatus.ID)
		if lb.ID == lbStatus.ID {
			i++
			fmt.Println("Targets: ", lbStatus.Targets)
			controlPlaneStatusIDs = lbStatus.Targets
			fmt.Println("controlPlaneStatusIDs: ", controlPlaneStatusIDs)
		}
	}

	if i == 0 {
		return nil, nil, fmt.Errorf("Could not find main load balancer in status - ControlPlaneLoadBalancers %s", "error")
	}

	fmt.Println("These are the control plane status ids: ", controlPlaneStatusIDs)

	var controlPlaneIDs intSlice

	for _, cp := range controlPlanes {

		controlPlaneIDs = append(controlPlaneIDs, cp.ID)

		// Check whether control plane is in target set of load balancer
		// If not than add it
		if !controlPlaneStatusIDs.contains(cp.ID) {
			fmt.Println("This ID gets added: ", cp.ID)
			needCreation = append(needCreation, cp)
		}
	}

	fmt.Println("These are the control plane IDs: ", controlPlaneIDs)

	// Check whether all the targets of the load balancer still exist
	for _, id := range controlPlaneStatusIDs {
		if !controlPlaneIDs.contains(id) {
			server, _, err := s.scope.HcloudClient().GetServerByID(ctx, id)
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

func (s *Service) listControlPlanes() ([]*hcloud.Server, error) {
	labels := map[string]string{
		infrav1.ClusterTagKey(s.scope.HcloudCluster.Name): string(infrav1.ResourceLifecycleOwned),
		"machine_type": "control_plane",
	}
	opts := hcloud.ServerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(labels)
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

	// If server has not been added to the network yet, then the load balancer cannot add it
	if len(server.PrivateNet) == 0 {
		return nil
	}

	myBool := true
	loadBalancerAddServerTargetOpts := hcloud.LoadBalancerAddServerTargetOpts{Server: server, UsePrivateIP: &myBool}

	loadBalancers, err := s.scope.HcloudClient().ListLoadBalancers(ctx, hcloud.LoadBalancerListOpts{})
	if err != nil {
		return err
	}
	// This only works if there is only one load balancer
	if len(loadBalancers) == 0 {
		return fmt.Errorf("There is no load balancer. Cannot add server %v", server.ID)
	}
	lb := loadBalancers[0]

	// If load balancer has not been attached to a network, then it cannot add a server
	if len(lb.PrivateNet) == 0 {
		return nil
	}
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
