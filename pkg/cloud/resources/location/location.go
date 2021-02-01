package location

import (
	"context"
	"fmt"
	"sort"
	"strings"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha4"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/scope"
	"github.com/hetznercloud/hcloud-go/hcloud"
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
	GetSpecLocations() []infrav1.HcloudLocation
	SetStatusLocations(locations []infrav1.HcloudLocation, networkZone infrav1.HcloudNetworkZone)
}

func (s *Service) Reconcile(ctx context.Context) (err error) {
	allLocations, err := s.scope.HcloudClient().ListLocation(ctx)
	if err != nil {
		return err
	}
	allLocationsMap := make(map[string]*hcloud.Location)
	for _, l := range allLocations {
		allLocationsMap[l.Name] = l
	}

	var locations []string
	var networkZone *infrav1.HcloudNetworkZone

	// if no locations have been specified, use the default networkZone
	specLocations := s.scope.GetSpecLocations()
	if len(specLocations) == 0 {
		nZ := infrav1.HcloudNetworkZone(hcloud.NetworkZoneEUCentral)
		networkZone = &nZ
		for _, l := range allLocations {
			if nZ == infrav1.HcloudNetworkZone(l.NetworkZone) {
				locations = append(locations, l.Name)
			}
		}
	}

	for _, l := range specLocations {
		location, ok := allLocationsMap[string(l)]
		if !ok {
			return fmt.Errorf("error location '%s' cannot be found", l)
		}
		nZ := infrav1.HcloudNetworkZone(location.NetworkZone)

		if networkZone == nil {
			networkZone = &nZ
		}

		if *networkZone != nZ {
			return fmt.Errorf(
				"error all locations need to be in the same NetworkZone. %s in NetworkZone %s, %s in NetworkZone %s",
				strings.Join(locations, ","),
				*networkZone,
				location.Name,
				nZ,
			)
		}
		locations = append(locations, location.Name)
	}

	if len(locations) == 0 {
		return fmt.Errorf("error locations is empty")
	}

	sort.Strings(locations)
	locationsTyped := make([]infrav1.HcloudLocation, len(locations))
	for pos := range locations {
		locationsTyped[pos] = infrav1.HcloudLocation(locations[pos])
	}

	s.scope.SetStatusLocations(locationsTyped, *networkZone)
	return nil
}
