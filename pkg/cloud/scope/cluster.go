package scope

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	HetznerClient
	Ctx                  context.Context
	HetznerClientFactory HetznerClientFactory
	Client               client.Client
	Logger               logr.Logger
	Cluster              *clusterv1.Cluster
	HetznerCluster       *infrav1.HetznerCluster
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("failed to generate new scope from nil Cluster")
	}
	if params.HetznerCluster == nil {
		return nil, errors.New("failed to generate new scope from nil HetznerCluster")
	}

	if params.Logger == nil {
		params.Logger = klogr.New()
	}

	if params.Ctx == nil {
		params.Ctx = context.TODO()
	}

	// setup client factory if nothing was set
	if params.HetznerClientFactory == nil {
		params.HetznerClientFactory = func(ctx context.Context) (HetznerClient, error) {
			// retrieve token secret
			var tokenSecret corev1.Secret
			tokenSecretName := types.NamespacedName{Namespace: params.HetznerCluster.Namespace, Name: params.HetznerCluster.Spec.TokenRef.Name}
			if err := params.Client.Get(ctx, tokenSecretName, &tokenSecret); err != nil {
				return nil, errors.Errorf("error getting referenced token secret/%s: %s", tokenSecretName, err)
			}

			tokenBytes, keyExists := tokenSecret.Data[params.HetznerCluster.Spec.TokenRef.Key]
			if !keyExists {
				return nil, errors.Errorf("error key %s does not exist in secret/%s", params.HetznerCluster.Spec.TokenRef.Key, tokenSecretName)
			}

			return &realClient{client: hcloud.NewClient(hcloud.WithToken(string(tokenBytes)))}, nil
		}
	}

	hc, err := params.HetznerClientFactory(params.Ctx)
	if err != nil {
		return nil, err
	}

	helper, err := patch.NewHelper(params.HetznerCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterScope{
		Ctx:            params.Ctx,
		Logger:         params.Logger,
		client:         params.Client,
		Cluster:        params.Cluster,
		HetznerCluster: params.HetznerCluster,
		hetznerClient:  hc,
		patchHelper:    helper,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	Ctx context.Context
	logr.Logger
	client        client.Client
	patchHelper   *patch.Helper
	hetznerClient HetznerClient

	Cluster        *clusterv1.Cluster
	HetznerCluster *infrav1.HetznerCluster
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HetznerCluster)
}

func (s *ClusterScope) HetznerClient() HetznerClient {
	return s.hetznerClient
}

func (s *ClusterScope) GetSpecLocation() infrav1.HetznerLocation {
	return s.HetznerCluster.Spec.Location
}

func (s *ClusterScope) SetStatusLocation(location infrav1.HetznerLocation, networkZone infrav1.HetznerNetworkZone) {
	s.HetznerCluster.Status.Location = location
	s.HetznerCluster.Status.NetworkZone = networkZone
}
