package scope

import (
	"context"
	"github.com/pkg/errors"
	bootstrapv1 "sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm/api/v1alpha2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	packerapi "github.com/simonswine/cluster-api-provider-hcloud/pkg/packer/api"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type MachineScopeParams struct {
	ClusterScopeParams
	KubeadmConfig *bootstrapv1.KubeadmConfig
	Machine       *clusterv1.Machine
	HcloudMachine *infrav1.HcloudMachine
}

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
		KubeadmConfig: params.KubeadmConfig,
		Machine:       params.Machine,
		HcloudMachine: params.HcloudMachine,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type MachineScope struct {
	ClusterScope
	KubeadmConfig *bootstrapv1.KubeadmConfig
	Machine       *clusterv1.Machine
	HcloudMachine *infrav1.HcloudMachine
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *MachineScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HcloudMachine)
}

func (s *MachineScope) GetSpecLocation() infrav1.HcloudLocation {
	return s.HcloudMachine.Spec.Location
}

func (s *MachineScope) SetStatusLocation(location infrav1.HcloudLocation, networkZone infrav1.HcloudNetworkZone) {
	s.HcloudMachine.Status.Location = location
	s.HcloudMachine.Status.NetworkZone = networkZone
}

func (s *MachineScope) EnsureImage(ctx context.Context, parameters *packerapi.PackerParameters) (*infrav1.HcloudImageID, error) {
	return s.packer.EnsureImage(ctx, s, s.hcloudClient, parameters)
}

func (s *MachineScope) IsControlPlane() bool {
	_, ok := s.Machine.Labels[clusterv1.MachineControlPlaneLabelName]
	return ok
}
