package scope

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "sigs.k8s.io/cluster-api-provider-hetzner/api/v1alpha3"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type VolumeScopeParams struct {
	ClusterScopeParams
	HetznerVolume *infrav1.HetznerVolume
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewVolumeScope(params VolumeScopeParams) (*VolumeScope, error) {
	if params.HetznerVolume == nil {
		return nil, errors.New("failed to generate new scope from nil HetznerVolume")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.HetznerVolume, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &VolumeScope{
		ClusterScope:  *cs,
		HetznerVolume: params.HetznerVolume,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type VolumeScope struct {
	ClusterScope
	HetznerVolume *infrav1.HetznerVolume
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *VolumeScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HetznerVolume)
}

func (s *VolumeScope) GetSpecLocation() infrav1.HetznerLocation {
	return s.HetznerVolume.Spec.Location
}

func (s *VolumeScope) SetStatusLocation(location infrav1.HetznerLocation, networkZone infrav1.HetznerNetworkZone) {
	s.HetznerVolume.Status.Location = location
}
