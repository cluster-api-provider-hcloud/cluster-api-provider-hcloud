package scope

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	packerapi "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer/api"
)

// BastionHostScopeParams defines the input parameters used to create a new Scope.
type BastionHostScopeParams struct {
	ClusterScopeParams
	BastionHost *infrav1.BastionHost
}

// NewBastionHostScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewBastionHostScope(params BastionHostScopeParams) (*BastionHostScope, error) {

	if params.BastionHost == nil {
		return nil, errors.New("failed to generate new scope from nil BastionHost")
	}

	cs, err := NewClusterScope(params.ClusterScopeParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cs.patchHelper, err = patch.NewHelper(params.BastionHost, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &BastionHostScope{
		ClusterScope: *cs,
		BastionHost:  params.BastionHost,
	}, nil
}

// BastionHostScope defines the basic context for an actuator to operate upon.
type BastionHostScope struct {
	ClusterScope
	BastionHost *infrav1.BastionHost
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *BastionHostScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.BastionHost)
}

func (s *BastionHostScope) EnsureImage(ctx context.Context, parameters *packerapi.PackerParameters) (*infrav1.HcloudImageID, error) {
	return s.packer.EnsureImage(ctx, s, s.hcloudClient, parameters)
}
