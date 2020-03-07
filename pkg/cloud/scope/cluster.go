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

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/manifests/parameters"
)

const defaultControlPlaneAPIEndpointPort = 6443

type Packer interface {
}

type Manifests interface {
	Apply(ctx context.Context, client clientcmd.ClientConfig, extVar map[string]string) error
}

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	HetznerClient
	Ctx                  context.Context
	HetznerClientFactory HetznerClientFactory
	Client               client.Client
	Logger               logr.Logger
	Cluster              *clusterv1.Cluster
	HetznerCluster       *infrav1.HetznerCluster
	Packer               Packer
	Manifests            Manifests
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
	var hetznerToken string
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
			hetznerToken = string(tokenBytes)

			return &realClient{client: hcloud.NewClient(hcloud.WithToken(hetznerToken))}, nil
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
		Client:         params.Client,
		Cluster:        params.Cluster,
		HetznerCluster: params.HetznerCluster,
		hetznerClient:  hc,
		hetznerToken:   hetznerToken,
		patchHelper:    helper,
		packer:         params.Packer,
		manifests:      params.Manifests,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	Ctx context.Context
	logr.Logger
	Client        client.Client
	patchHelper   *patch.Helper
	hetznerClient HetznerClient
	hetznerToken  string
	packer        Packer
	manifests     Manifests

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

	for _, floatingIP := range s.HetznerCluster.Status.ControlPlaneFloatingIPs {
		p.HcloudFloatingIPs = append(
			p.HcloudFloatingIPs,
			floatingIP.IP,
		)
	}

	p.HcloudToken = &s.hetznerToken

	if s.HetznerCluster.Status.Network == nil {
		return nil, errors.New("No network found")
	}
	hcloudNetwork := intstr.FromInt(s.HetznerCluster.Status.Network.ID)
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
