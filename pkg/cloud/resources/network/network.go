package network

import (
	"context"
	"net"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/utils"
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

func apiToStatus(network *hcloud.Network) *infrav1.HetznerNetworkStatus {
	var subnets = make([]infrav1.HetznerNetworkSubnetSpec, len(network.Subnets))
	for pos, n := range network.Subnets {
		subnets[pos].NetworkZone = infrav1.HetznerNetworkZone(n.NetworkZone)
		subnets[pos].CIDRBlock = n.IPRange.String()
	}

	var status infrav1.HetznerNetworkStatus
	status.ID = network.ID
	status.CIDRBlock = network.IPRange.String()
	status.Subnets = subnets
	status.Labels = network.Labels
	return &status
}

func (s *Service) defaults() *infrav1.HetznerNetworkSpec {
	n := infrav1.HetznerNetworkSpec{}
	n.CIDRBlock = "10.0.0.0/16"
	n.Subnets = []infrav1.HetznerNetworkSubnetSpec{
		{
			NetworkZone: s.scope.HetznerCluster.Status.NetworkZone,
			HetznerNetwork: infrav1.HetznerNetwork{
				CIDRBlock: "10.0.0.0/24",
			},
		},
	}
	return &n
}

func (s *Service) labels() map[string]string {
	clusterTagKey := infrav1.ClusterTagKey(s.scope.HetznerCluster.Name)
	return map[string]string{
		clusterTagKey: string(infrav1.ResourceLifecycleOwned),
	}
}

func (s *Service) Reconcile(ctx context.Context) (err error) {
	// update current status
	networkStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh networks")
	}
	s.scope.HetznerCluster.Status.Network = networkStatus

	if networkStatus != nil {
		// TODO: Check if the actual values are matching
		return nil
	}

	if s.scope.HetznerCluster.Spec.Network == nil {
		s.scope.HetznerCluster.Spec.Network = s.defaults()
	}

	networkStatus, err = s.createNetwork(ctx, s.scope.HetznerCluster.Spec.Network)
	if err != nil {
		return errors.Wrap(err, "failed to create network")
	}
	s.scope.HetznerCluster.Status.Network = networkStatus

	return nil
}

func (s *Service) createNetwork(ctx context.Context, spec *infrav1.HetznerNetworkSpec) (*infrav1.HetznerNetworkStatus, error) {
	hc := s.scope.HetznerCluster

	s.scope.V(2).Info("Create a new network", "cidrBlock", spec.CIDRBlock, "subnets", spec.Subnets)
	_, network, err := net.ParseCIDR(spec.CIDRBlock)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid network '%s'", spec.CIDRBlock)
	}

	var subnets = make([]hcloud.NetworkSubnet, len(spec.Subnets))
	for pos, sn := range spec.Subnets {
		_, subnet, err := net.ParseCIDR(sn.CIDRBlock)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid network '%s'", sn.CIDRBlock)
		}
		subnets[pos].IPRange = subnet
		subnets[pos].NetworkZone = hcloud.NetworkZone(sn.NetworkZone)
		subnets[pos].Type = hcloud.NetworkSubnetTypeServer
	}

	opts := hcloud.NetworkCreateOpts{
		Name:    hc.Name,
		IPRange: network,
		Labels:  s.labels(),
		Subnets: subnets,
	}

	s.scope.V(1).Info("Create a new network", "opts", opts)

	respNetworkCreate, _, err := s.scope.HetznerClient().CreateNetwork(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "error creating network")
	}

	return apiToStatus(respNetworkCreate), nil
}

func (s *Service) deleteNetwork(ctx context.Context, status *infrav1.HetznerNetworkStatus) error {
	// ensure deleted network is actually owned by us
	clusterTagKey := infrav1.ClusterTagKey(s.scope.HetznerCluster.Name)
	if status.Labels == nil || infrav1.ResourceLifecycle(status.Labels[clusterTagKey]) != infrav1.ResourceLifecycleOwned {
		s.scope.V(3).Info("Ignore request to delete network, as it is not owned", "id", status.ID, "cidrBlock", status.CIDRBlock)
		return nil
	}
	_, err := s.scope.HetznerClient().DeleteNetwork(ctx, &hcloud.Network{ID: status.ID})
	s.scope.V(2).Info("Delete network", "id", status.ID, "cidrBlock", status.CIDRBlock)
	return err
}

func (s *Service) Delete(ctx context.Context) (err error) {
	// update current status
	networkStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh networks")
	}

	if err := s.deleteNetwork(ctx, networkStatus); err != nil {
		return errors.Wrap(err, "failed to delete network")
	}

	return nil
}

func (s *Service) actualStatus(ctx context.Context) (*infrav1.HetznerNetworkStatus, error) {
	opts := hcloud.NetworkListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.labels())
	networks, err := s.scope.HetznerClient().ListNetworks(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	for _, n := range networks {
		return apiToStatus(n), nil
	}

	return nil, nil
}
