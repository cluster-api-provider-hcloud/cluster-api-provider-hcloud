package scope

import (
	"context"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/util/patch"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type BareMetalMachineScopeParams struct {
	ClusterScopeParams
	BareMetalMachine *infrav1.BareMetalMachine
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewBareMetalMachineScope(params BareMetalMachineScopeParams) (*BareMetalMachineScope, error) {
	if params.BareMetalMachine == nil {
		return nil, errors.New("failed to generate new scope from nil BareMetalMachine")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.BareMetalMachine, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &BareMetalMachineScope{
		ClusterScope:     *cs,
		BareMetalMachine: params.BareMetalMachine,
	}, nil
}

// BareMetalMachineScope defines the basic context for an actuator to operate upon.
type BareMetalMachineScope struct {
	ClusterScope
	BareMetalMachine *infrav1.BareMetalMachine
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *BareMetalMachineScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.BareMetalMachine)
}

// Name returns the BareMetalMachine name.
func (s *BareMetalMachineScope) Name() string {
	return s.BareMetalMachine.Name
}

// Namespace returns the namespace name.
func (s *BareMetalMachineScope) Namespace() string {
	return s.BareMetalMachine.Namespace
}

// IP returns the IP address
func (s *BareMetalMachineScope) IP() string {
	return *s.BareMetalMachine.Spec.IP
}

// PatchObject persists the machine spec and status.
func (s *BareMetalMachineScope) PatchObject(ctx context.Context) error {
	return s.patchHelper.Patch(ctx, s.BareMetalMachine)
}
