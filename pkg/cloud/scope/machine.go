package scope

import (
	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type MachineScopeParams struct {
	ClusterScopeParams
	Machine        *clusterv1.Machine
	HetznerMachine *infrav1.HetznerMachine
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewMachineScope(params MachineScopeParams) (*MachineScope, error) {
	if params.Machine == nil {
		return nil, errors.New("failed to generate new scope from nil Machine")
	}
	if params.HetznerMachine == nil {
		return nil, errors.New("failed to generate new scope from nil HetznerMachine")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.HetznerMachine, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &MachineScope{
		ClusterScope:   *cs,
		Machine:        params.Machine,
		HetznerMachine: params.HetznerMachine,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type MachineScope struct {
	ClusterScope
	Machine        *clusterv1.Machine
	HetznerMachine *infrav1.HetznerMachine
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *MachineScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HetznerMachine)
}

func (s *MachineScope) GetSpecLocation() infrav1.HetznerLocation {
	return s.HetznerMachine.Spec.Location
}

func (s *MachineScope) SetStatusLocation(location infrav1.HetznerLocation, networkZone infrav1.HetznerNetworkZone) {
	s.HetznerMachine.Status.Location = location
	s.HetznerMachine.Status.NetworkZone = networkZone
}
