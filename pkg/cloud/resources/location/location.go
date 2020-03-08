package location

import (
	"context"
	"fmt"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/scope"
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
	HcloudClient() scope.HcloudClient
	GetSpecLocation() infrav1.HcloudLocation
	SetStatusLocation(location infrav1.HcloudLocation, networkZone infrav1.HcloudNetworkZone)
}

func (s *Service) Reconcile(ctx context.Context) (err error) {
	locations, err := s.scope.HcloudClient().ListLocation(ctx)
	if err != nil {
		return err
	}

	specLocation := s.scope.GetSpecLocation()
	for _, location := range locations {
		if location.Name == string(specLocation) {
			s.scope.SetStatusLocation(specLocation, infrav1.HcloudNetworkZone(location.NetworkZone))
			return nil
		}
	}
	return fmt.Errorf("error location '%s' cannot be found", specLocation)
}
