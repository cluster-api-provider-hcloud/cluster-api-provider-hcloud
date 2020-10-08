package scope

import (
	"context"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util/patch"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type BareMetalMachineScopeParams struct {
	ClusterScopeParams
	Machine          *clusterv1.Machine
	BareMetalMachine *infrav1.BareMetalMachine
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewBareMetalMachineScope(params BareMetalMachineScopeParams) (*BareMetalMachineScope, error) {
	if params.Machine == nil {
		return nil, errors.New("failed to generate new scope from nil Machine")
	}
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
		Machine:          params.Machine,
		BareMetalMachine: params.BareMetalMachine,
	}, nil
}

// BareMetalMachineScope defines the basic context for an actuator to operate upon.
type BareMetalMachineScope struct {
	ClusterScope
	Machine          *clusterv1.Machine
	BareMetalMachine *infrav1.BareMetalMachine
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *BareMetalMachineScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.BareMetalMachine)
}

// Name returns the BareMetalMachine name.
func (m *BareMetalMachineScope) Name() string {
	return m.BareMetalMachine.Name
}

// Namespace returns the namespace name.
func (m *BareMetalMachineScope) Namespace() string {
	return m.BareMetalMachine.Namespace
}

// PatchObject persists the machine spec and status.
func (m *BareMetalMachineScope) PatchObject(ctx context.Context) error {
	return m.patchHelper.Patch(ctx, m.BareMetalMachine)
}

func (m *BareMetalMachineScope) SetFailureReason(reason capierrors.MachineStatusError) {
	m.BareMetalMachine.Status.FailureReason = &reason
}

func (m *BareMetalMachineScope) SetFailureMessage(err error) {
	m.BareMetalMachine.Status.FailureMessage = pointer.StringPtr(err.Error())
}

// GetRawBootstrapData returns the bootstrap data from the secret in the BareMetalMachine's bootstrap.dataSecretName.
func (m *BareMetalMachineScope) GetRawBootstrapData(ctx context.Context) ([]byte, error) {
	if m.Machine.Spec.Bootstrap.DataSecretName == nil {
		return nil, ErrBootstrapDataNotReady
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: m.Namespace(), Name: *m.Machine.Spec.Bootstrap.DataSecretName}
	if err := m.Client.Get(ctx, key, secret); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve bootstrap data secret for BareMetalMachine %s/%s", m.Namespace(), m.Name())
	}

	value, ok := secret.Data["value"]
	if !ok {
		return nil, errors.New("error retrieving bootstrap data: secret value key is missing")
	}

	return value, nil
}
