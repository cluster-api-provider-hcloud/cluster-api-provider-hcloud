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
type HetznerBareMetalMachineScopeParams struct {
	ClusterScopeParams
	Machine                 *clusterv1.Machine
	HetznerBareMetalMachine *infrav1.HetznerBareMetalMachine
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewHetznerBareMetalMachineScope(params HetznerBareMetalMachineScopeParams) (*HetznerBareMetalMachineScope, error) {
	if params.Machine == nil {
		return nil, errors.New("failed to generate new scope from nil Machine")
	}
	if params.HetznerBareMetalMachine == nil {
		return nil, errors.New("failed to generate new scope from nil HetznerBareMetalMachine")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.HetznerBareMetalMachine, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &HetznerBareMetalMachineScope{
		ClusterScope:            *cs,
		Machine:                 params.Machine,
		HetznerBareMetalMachine: params.HetznerBareMetalMachine,
	}, nil
}

// HetznerBareMetalMachineScope defines the basic context for an actuator to operate upon.
type HetznerBareMetalMachineScope struct {
	ClusterScope
	Machine                 *clusterv1.Machine
	HetznerBareMetalMachine *infrav1.HetznerBareMetalMachine
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *HetznerBareMetalMachineScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HetznerBareMetalMachine)
}

// Name returns the BareMetalMachine name.
func (m *HetznerBareMetalMachineScope) Name() string {
	return m.HetznerBareMetalMachine.Name
}

// Namespace returns the namespace name.
func (m *HetznerBareMetalMachineScope) Namespace() string {
	return m.HetznerBareMetalMachine.Namespace
}

// PatchObject persists the machine spec and status.
func (m *HetznerBareMetalMachineScope) PatchObject(ctx context.Context) error {
	return m.patchHelper.Patch(ctx, m.HetznerBareMetalMachine)
}

func (m *HetznerBareMetalMachineScope) SetFailureReason(reason capierrors.MachineStatusError) {
	m.HetznerBareMetalMachine.Status.FailureReason = &reason
}

func (m *HetznerBareMetalMachineScope) SetFailureMessage(err error) {
	m.HetznerBareMetalMachine.Status.FailureMessage = pointer.StringPtr(err.Error())
}

func (m *HetznerBareMetalMachineScope) IsBootstrapDataReady(ctx context.Context) bool {
	return m.Machine.Spec.Bootstrap.DataSecretName != nil
}

// GetRawBootstrapData returns the bootstrap data from the secret in the HetznerBareMetalMachine's bootstrap.dataSecretName.
func (m *HetznerBareMetalMachineScope) GetRawBootstrapData(ctx context.Context) ([]byte, error) {
	if m.Machine.Spec.Bootstrap.DataSecretName == nil {
		return nil, ErrBootstrapDataNotReady
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: m.Namespace(), Name: *m.Machine.Spec.Bootstrap.DataSecretName}
	if err := m.Client.Get(ctx, key, secret); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve bootstrap data secret for HetznerBareMetalMachine %s/%s", m.Namespace(), m.Name())
	}

	value, ok := secret.Data["value"]
	if !ok {
		return nil, errors.New("error retrieving bootstrap data: secret value key is missing")
	}

	return value, nil
}
