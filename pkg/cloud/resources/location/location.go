package location

import (
	"context"
	"fmt"

	infrav1 "sigs.k8s.io/cluster-api-provider-hetzner/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-hetzner/pkg/cloud/scope"
)

type Service struct {
	scope localScope
}

func NewService(scope localScope) *Service {
	return &Service{
		scope: scope,
	}
}

type localScope interface {
	HetznerClient() scope.HetznerClient
	GetSpecLocation() infrav1.HetznerLocation
	SetStatusLocation(location infrav1.HetznerLocation, networkZone infrav1.HetznerNetworkZone)
}

func (s *Service) Reconcile(ctx context.Context) (err error) {
	locations, err := s.scope.HetznerClient().ListLocation(ctx)
	if err != nil {
		return err
	}

	specLocation := s.scope.GetSpecLocation()
	for _, location := range locations {
		if location.Name == string(specLocation) {
			s.scope.SetStatusLocation(specLocation, infrav1.HetznerNetworkZone(location.NetworkZone))
			return nil
		}
	}
	return fmt.Errorf("error location '%s' cannot be found", specLocation)
}
