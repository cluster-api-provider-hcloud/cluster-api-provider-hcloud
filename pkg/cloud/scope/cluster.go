package scope

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/manifests/parameters"
)

const defaultControlPlaneAPIEndpointPort = 6443

type Packer interface {
}

type Manifests interface {
	Apply(ctx context.Context, client clientcmd.ClientConfig, extVar map[string]string) error
}

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	HcloudClient
	Ctx                 context.Context
	HcloudClientFactory HcloudClientFactory
	Client              client.Client
	Logger              logr.Logger
	Cluster             *clusterv1.Cluster
	HcloudCluster       *infrav1.HcloudCluster
	Packer              Packer
	Manifests           Manifests
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("failed to generate new scope from nil Cluster")
	}
	if params.HcloudCluster == nil {
		return nil, errors.New("failed to generate new scope from nil HcloudCluster")
	}
	if params.Packer == nil {
		return nil, errors.New("failed to generate new scope from nil Packer")
	}
	if params.Manifests == nil {
		return nil, errors.New("failed to generate new scope from nil Manifests")
	}

	if params.Logger == nil {
		params.Logger = klogr.New()
	}

	if params.Ctx == nil {
		params.Ctx = context.TODO()
	}

	// setup client factory if nothing was set
	var hcloudToken string
	if params.HcloudClientFactory == nil {
		params.HcloudClientFactory = func(ctx context.Context) (HcloudClient, error) {
			// retrieve token secret
			var tokenSecret corev1.Secret
			tokenSecretName := types.NamespacedName{Namespace: params.HcloudCluster.Namespace, Name: params.HcloudCluster.Spec.TokenRef.Name}
			if err := params.Client.Get(ctx, tokenSecretName, &tokenSecret); err != nil {
				return nil, errors.Errorf("error getting referenced token secret/%s: %s", tokenSecretName, err)
			}

			tokenBytes, keyExists := tokenSecret.Data[params.HcloudCluster.Spec.TokenRef.Key]
			if !keyExists {
				return nil, errors.Errorf("error key %s does not exist in secret/%s", params.HcloudCluster.Spec.TokenRef.Key, tokenSecretName)
			}
			hcloudToken = string(tokenBytes)

			return &realClient{client: hcloud.NewClient(hcloud.WithToken(hcloudToken))}, nil
		}
	}

	hc, err := params.HcloudClientFactory(params.Ctx)
	if err != nil {
		return nil, err
	}

	helper, err := patch.NewHelper(params.HcloudCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterScope{
		Ctx:           params.Ctx,
		Logger:        params.Logger,
		Client:        params.Client,
		Cluster:       params.Cluster,
		HcloudCluster: params.HcloudCluster,
		hcloudClient:  hc,
		hcloudToken:   hcloudToken,
		patchHelper:   helper,
		packer:        params.Packer,
		manifests:     params.Manifests,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	Ctx context.Context
	logr.Logger
	Client       client.Client
	patchHelper  *patch.Helper
	hcloudClient HcloudClient
	hcloudToken  string
	packer       Packer
	manifests    Manifests

	Cluster       *clusterv1.Cluster
	HcloudCluster *infrav1.HcloudCluster
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HcloudCluster)
}

func (s *ClusterScope) HcloudClient() HcloudClient {
	return s.hcloudClient
}

func (s *ClusterScope) GetSpecLocation() infrav1.HcloudLocation {
	return s.HcloudCluster.Spec.Location
}

func (s *ClusterScope) SetStatusLocation(location infrav1.HcloudLocation, networkZone infrav1.HcloudNetworkZone) {
	s.HcloudCluster.Status.Location = location
	s.HcloudCluster.Status.NetworkZone = networkZone
}

func (s *ClusterScope) ControlPlaneAPIEndpointPort() int {
	return defaultControlPlaneAPIEndpointPort
}

// ClientConfig return a kubernetes client config for the cluster context
func (s *ClusterScope) ClientConfig() (clientcmd.ClientConfig, error) {
	kubeconfigBytes, err := kubeconfig.FromSecret(s.Client, s.Cluster)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving kubeconfig for cluster")
	}
	return clientcmd.NewClientConfigFromBytes(kubeconfigBytes)
}

func (s *ClusterScope) ClientConfigWithAPIEndpoint(endpoint clusterv1.APIEndpoint) (clientcmd.ClientConfig, error) {
	c, err := s.ClientConfig()
	if err != nil {
		return nil, err
	}

	raw, err := c.RawConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving rawConfig from clientConfig")
	}

	// update cluster endpint in confgi
	for key := range raw.Clusters {
		raw.Clusters[key].Server = fmt.Sprintf("https://%s:%d", endpoint.Host, endpoint.Port)
	}

	return clientcmd.NewDefaultClientConfig(raw, &clientcmd.ConfigOverrides{}), nil
}

func (s *ClusterScope) manifestParameters() (*parameters.ManifestParameters, error) {
	var p parameters.ManifestParameters

	for _, floatingIP := range s.HcloudCluster.Status.ControlPlaneFloatingIPs {
		p.HcloudFloatingIPs = append(
			p.HcloudFloatingIPs,
			floatingIP.IP,
		)
	}

	p.HcloudToken = &s.hcloudToken

	if s.HcloudCluster.Status.Network == nil {
		return nil, errors.New("No network found")
	}
	hcloudNetwork := intstr.FromInt(s.HcloudCluster.Status.Network.ID)
	p.HcloudNetwork = &hcloudNetwork

	if s.Cluster.Spec.ClusterNetwork == nil || s.Cluster.Spec.ClusterNetwork.Pods == nil || len(s.Cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) == 0 {
		return nil, errors.New("No pod cidr network set")
	}
	if len(s.Cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) != 1 {
		return nil, errors.New("More than one pod cidr network set")
	}
	_, podCidrBlock, err := net.ParseCIDR(s.Cluster.Spec.ClusterNetwork.Pods.CIDRBlocks[0])
	if err != nil {
		return nil, errors.Wrap(err, "Invalid pod cidr network set")
	}
	p.PodCIDRBlock = podCidrBlock

	return &p, nil
}

func (s *ClusterScope) ApplyManifestsWithClientConfig(ctx context.Context, c clientcmd.ClientConfig) error {
	manifestParameters, err := s.manifestParameters()
	if err != nil {
		return err
	}

	return s.manifests.Apply(ctx, c, manifestParameters.ExtVar())
}
