package scope

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type VolumeScopeParams struct {
	ClusterScopeParams
	HcloudVolume *infrav1.HcloudVolume
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewVolumeScope(params VolumeScopeParams) (*VolumeScope, error) {
	if params.HcloudVolume == nil {
		return nil, errors.New("failed to generate new scope from nil HcloudVolume")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.HcloudVolume, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &VolumeScope{
		ClusterScope: *cs,
		HcloudVolume: params.HcloudVolume,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type VolumeScope struct {
	ClusterScope
	HcloudVolume *infrav1.HcloudVolume
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *VolumeScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HcloudVolume)
}

func (s *VolumeScope) GetSpecLocations() []infrav1.HcloudLocation {
	return []infrav1.HcloudLocation{s.HcloudVolume.Spec.Location}
}

func (s *VolumeScope) SetStatusLocations(location []infrav1.HcloudLocation, networkZone infrav1.HcloudNetworkZone) {
	s.HcloudVolume.Status.Location = location[0]
}
