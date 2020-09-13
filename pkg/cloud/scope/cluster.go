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
	"k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests/parameters"
	packerapi "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer/api"
)

const defaultControlPlaneAPIEndpointPort = 6443

type Packer interface {
	EnsureImage(ctx context.Context, log logr.Logger, hc packerapi.HcloudClient, parameters *packerapi.PackerParameters) (*infrav1.HcloudImageID, error)
}

type Manifests interface {
	Apply(ctx context.Context, client clientcmd.ClientConfig, extVar map[string]string) error
	Hash(extVar map[string]string) (string, error)
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

			return &realClient{client: hcloud.NewClient(hcloud.WithToken(hcloudToken)), token: hcloudToken}, nil
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

// Name returns the HcloudCluster name.
func (m *ClusterScope) Name() string {
	return m.HcloudCluster.Name
}

// Namespace returns the namespace name.
func (m *ClusterScope) Namespace() string {
	return m.HcloudCluster.Namespace
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close() error {
	return s.patchHelper.Patch(s.Ctx, s.HcloudCluster)
}

func (s *ClusterScope) HcloudClient() HcloudClient {
	return s.hcloudClient
}

func (s *ClusterScope) GetSpecLocations() []infrav1.HcloudLocation {
	return s.HcloudCluster.Spec.Locations
}

func (s *ClusterScope) SetStatusLocations(locations []infrav1.HcloudLocation, networkZone infrav1.HcloudNetworkZone) {
	s.HcloudCluster.Spec.Locations = locations
	s.HcloudCluster.Status.Locations = locations
	s.HcloudCluster.Status.FailureDomains = make(clusterv1.FailureDomains)
	for _, l := range locations {
		s.HcloudCluster.Status.FailureDomains[string(l)] = clusterv1.FailureDomainSpec{
			ControlPlane: true,
		}
	}
	s.HcloudCluster.Status.NetworkZone = networkZone
}

func (s *ClusterScope) ControlPlaneAPIEndpointPort() int32 {
	return defaultControlPlaneAPIEndpointPort
}

// ClientConfig return a kubernetes client config for the cluster context
func (s *ClusterScope) ClientConfig() (clientcmd.ClientConfig, error) {
	var cluster = client.ObjectKey{
		Name:      s.Cluster.Name,
		Namespace: s.Cluster.Namespace,
	}
	kubeconfigBytes, err := kubeconfig.FromSecret(s.Ctx, s.Client, cluster)
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

	p.KubeAPIServerIPv4 = &s.HcloudCluster.Status.ControlPlaneLoadBalancers[0].IPv4

	p.KubeAPIServerDomain = &s.HcloudCluster.Status.KubeAPIServerDomain

	p.HcloudToken = &s.hcloudToken

	if s.HcloudCluster.Status.Network != nil {
		hcloudNetwork := intstr.FromInt(s.HcloudCluster.Status.Network.ID)
		p.HcloudNetwork = &hcloudNetwork
	}

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

	p.Manifests = s.HcloudCluster.Spec.Manifests
	return &p, nil
}

func (s *ClusterScope) ListMachines(ctx context.Context) ([]*clusterv1.Machine, []*infrav1.HcloudMachine, error) {
	// get and index Machines by HcloudMachine name
	var machineListRaw clusterv1.MachineList
	var machineByHcloudMachineName = make(map[string]*clusterv1.Machine)
	if err := s.Client.List(ctx, &machineListRaw, client.InNamespace(s.Namespace())); err != nil {
		return nil, nil, err
	}
	expectedGK := infrav1.GroupVersion.WithKind("HcloudMachine").GroupKind()
	for pos := range machineListRaw.Items {
		m := &machineListRaw.Items[pos]
		actualGK := m.Spec.InfrastructureRef.GroupVersionKind().GroupKind()
		if m.Spec.ClusterName != s.Cluster.Name ||
			actualGK.String() != expectedGK.String() {
			continue
		}
		machineByHcloudMachineName[m.Spec.InfrastructureRef.Name] = m
	}

	// match HcloudMachines to Machines
	var hcloudMachineListRaw infrav1.HcloudMachineList
	if err := s.Client.List(ctx, &hcloudMachineListRaw, client.InNamespace(s.Namespace())); err != nil {
		return nil, nil, err
	}
	var machineList []*clusterv1.Machine
	var hcloudMachineList []*infrav1.HcloudMachine
	for pos := range hcloudMachineListRaw.Items {
		hm := &hcloudMachineListRaw.Items[pos]
		m, ok := machineByHcloudMachineName[hm.Name]
		if !ok {
			continue
		}

		machineList = append(machineList, m)
		hcloudMachineList = append(hcloudMachineList, hm)
	}

	return machineList, hcloudMachineList, nil
}

func IsControlPlaneReady(ctx context.Context, c clientcmd.ClientConfig) error {
	restConfig, err := c.ClientConfig()
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	_, err = clientSet.Discovery().RESTClient().Get().AbsPath("/readyz").DoRaw(ctx)
	return err
}

func (s *ClusterScope) ManifestsHash() (string, error) {
	manifestParameters, err := s.manifestParameters()
	if err != nil {
		return "", err
	}
	return s.manifests.Hash(manifestParameters.ExtVar())
}

func (s *ClusterScope) ApplyManifestsWithClientConfig(ctx context.Context, c clientcmd.ClientConfig) error {
	manifestParameters, err := s.manifestParameters()
	if err != nil {
		return err
	}
	return s.manifests.Apply(ctx, c, manifestParameters.ExtVar())
}
