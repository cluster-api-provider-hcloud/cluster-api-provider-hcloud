package scope

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	packerapi "github.com/simonswine/cluster-api-provider-hcloud/pkg/packer/api"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type MachineScopeParams struct {
	ClusterScopeParams
	Machine       *clusterv1.Machine
	HcloudMachine *infrav1.HcloudMachine
}

var ErrBootstrapDataNotReady = errors.New("error retrieving bootstrap data: linked Machine's bootstrap.dataSecretName is nil")

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewMachineScope(params MachineScopeParams) (*MachineScope, error) {
	if params.Machine == nil {
		return nil, errors.New("failed to generate new scope from nil Machine")
	}
	if params.HcloudMachine == nil {
		return nil, errors.New("failed to generate new scope from nil HcloudMachine")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.HcloudMachine, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &MachineScope{
		ClusterScope:  *cs,
		Machine:       params.Machine,
		HcloudMachine: params.HcloudMachine,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type MachineScope struct {
	ClusterScope
	Machine       *clusterv1.Machine
	HcloudMachine *infrav1.HcloudMachine
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *MachineScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HcloudMachine)
}

func (s *MachineScope) EnsureImage(ctx context.Context, parameters *packerapi.PackerParameters) (*infrav1.HcloudImageID, error) {
	return s.packer.EnsureImage(ctx, s, s.hcloudClient, parameters)
}

// IsControlPlane returns true if the machine is a control plane.
func (m *MachineScope) IsControlPlane() bool {
	return util.IsControlPlaneMachine(m.Machine)
}

// Name returns the HcloudMachine name.
func (m *MachineScope) Name() string {
	return m.HcloudMachine.Name
}

// Namespace returns the namespace name.
func (m *MachineScope) Namespace() string {
	return m.HcloudMachine.Namespace
}

// PatchObject persists the machine spec and status.
func (m *MachineScope) PatchObject(ctx context.Context) error {
	return m.patchHelper.Patch(ctx, m.HcloudMachine)
}

// SetReady sets the HcloudMachine Ready Status
func (m *MachineScope) SetReady() {
	m.HcloudMachine.Status.Ready = true
}

// SetNotReady sets the HcloudMachine Ready Status to false
func (m *MachineScope) SetNotReady() {
	m.HcloudMachine.Status.Ready = false
}

// SetFailureMessage sets the HcloudMachine status failure message.
func (m *MachineScope) SetFailureMessage(v error) {
	m.HcloudMachine.Status.FailureMessage = pointer.StringPtr(v.Error())
}

// SetFailureReason sets the HcloudMachine status failure reason.
func (m *MachineScope) SetFailureReason(v capierrors.MachineStatusError) {
	m.HcloudMachine.Status.FailureReason = &v
}

// GetRawBootstrapData returns the bootstrap data from the secret in the Machine's bootstrap.dataSecretName.
func (m *MachineScope) GetRawBootstrapData(ctx context.Context) ([]byte, error) {
	if m.Machine.Spec.Bootstrap.DataSecretName == nil {
		return nil, ErrBootstrapDataNotReady
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: m.Namespace(), Name: *m.Machine.Spec.Bootstrap.DataSecretName}
	if err := m.Client.Get(ctx, key, secret); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve bootstrap data secret for HcloudMachine %s/%s", m.Namespace(), m.Name())
	}

	value, ok := secret.Data["value"]
	if !ok {
		return nil, errors.New("error retrieving bootstrap data: secret value key is missing")
	}

	return value, nil
}

// HasFailed returns if a machine has failed permanently
func (m *MachineScope) HasFailed() bool {
	return m.HcloudMachine.Status.FailureReason != nil || m.HcloudMachine.Status.FailureMessage != nil
}
