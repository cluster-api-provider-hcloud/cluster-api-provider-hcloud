package scope

import (
	"context"
	"fmt"
	"strconv"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha4"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests/parameters"
	packerapi "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer/api"
	"github.com/go-logr/logr"
	"github.com/hetznercloud/hcloud-go/hcloud"
	hrobot "github.com/nl2go/hrobot-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	HrobotClient
	Ctx                 context.Context
	HcloudClientFactory HcloudClientFactory
	HrobotClientFactory HrobotClientFactory
	Client              client.Client
	Logger              logr.Logger
	Recorder            record.EventRecorder
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
			tokenSecretName := types.NamespacedName{Namespace: params.HcloudCluster.Namespace, Name: params.HcloudCluster.Spec.HcloudTokenRef.Name}
			if err := params.Client.Get(ctx, tokenSecretName, &tokenSecret); err != nil {
				return nil, errors.Errorf("error getting referenced token secret/%s: %s", tokenSecretName, err)
			}

			tokenBytes, keyExists := tokenSecret.Data[params.HcloudCluster.Spec.HcloudTokenRef.Key]
			if !keyExists {
				return nil, errors.Errorf("error key %s does not exist in secret/%s", params.HcloudCluster.Spec.HcloudTokenRef.Key, tokenSecretName)
			}
			hcloudToken = string(tokenBytes)

			return &realHcloudClient{client: hcloud.NewClient(hcloud.WithToken(hcloudToken)), token: hcloudToken}, nil
		}
	}

	var robotUserName string
	var robotPassword string
	var hrc HrobotClient
	if params.HcloudCluster.Spec.HrobotTokenRef != nil {
		if params.HrobotClientFactory == nil {
			params.HrobotClientFactory = func(ctx context.Context) (HrobotClient, error) {
				// retrieve token secret
				var tokenSecret corev1.Secret
				tokenSecretName := types.NamespacedName{Namespace: params.HcloudCluster.Namespace, Name: params.HcloudCluster.Spec.HrobotTokenRef.TokenName}
				if err := params.Client.Get(ctx, tokenSecretName, &tokenSecret); err != nil {
					return nil, errors.Errorf("error getting referenced token secret/%s: %s", tokenSecretName, err)
				}

				passwordTokenBytes, keyExists := tokenSecret.Data[params.HcloudCluster.Spec.HrobotTokenRef.PasswordKey]
				if !keyExists {
					return nil, errors.Errorf("error key %s does not exist in secret/%s", params.HcloudCluster.Spec.HrobotTokenRef.PasswordKey, tokenSecretName)
				}
				userNameTokenBytes, keyExists := tokenSecret.Data[params.HcloudCluster.Spec.HrobotTokenRef.UserNameKey]
				if !keyExists {
					return nil, errors.Errorf("error key %s does not exist in secret/%s", params.HcloudCluster.Spec.HrobotTokenRef.UserNameKey, tokenSecretName)
				}
				robotUserName = string(userNameTokenBytes)
				robotPassword = string(passwordTokenBytes)

				return &realHrobotClient{client: hrobot.NewBasicAuthClient(robotUserName, robotPassword)}, nil
			}
		}
		var err error
		hrc, err = params.HrobotClientFactory(params.Ctx)
		if err != nil {
			return nil, err
		}
	}

	hcc, err := params.HcloudClientFactory(params.Ctx)
	if err != nil {
		return nil, err
	}
	helper, err := patch.NewHelper(params.HcloudCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterScope{
		Ctx:           params.Ctx,
		Recorder:      params.Recorder,
		Logger:        params.Logger,
		Client:        params.Client,
		Cluster:       params.Cluster,
		HcloudCluster: params.HcloudCluster,
		hcloudClient:  hcc,
		hrobotClient:  hrc,
		hcloudToken:   hcloudToken,
		robotUserName: robotUserName,
		robotPassword: robotPassword,
		patchHelper:   helper,
		packer:        params.Packer,
		manifests:     params.Manifests,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	Ctx context.Context
	logr.Logger
	Recorder      record.EventRecorder
	Client        client.Client
	patchHelper   *patch.Helper
	hcloudClient  HcloudClient
	hrobotClient  HrobotClient
	hcloudToken   string
	robotUserName string
	robotPassword string
	packer        Packer
	manifests     Manifests

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

func (s *ClusterScope) HrobotClient() HrobotClient {
	return s.hrobotClient
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
		raw.Clusters[key].Server = fmt.Sprintf("https://%s:%d", s.HcloudCluster.Spec.ControlPlaneEndpoint.Host, s.HcloudCluster.Spec.ControlPlaneLoadBalancer.Services[0].ListenPort)
	}

	return clientcmd.NewDefaultClientConfig(raw, &clientcmd.ConfigOverrides{}), nil
}

func (s *ClusterScope) manifestParameters() (*parameters.ManifestParameters, error) {
	var p parameters.ManifestParameters

	p.KubeAPIServerIPv4 = &s.HcloudCluster.Status.ControlPlaneLoadBalancer.IPv4
	var emptyString = ""
	if s.HcloudCluster.Spec.ControlPlaneEndpoint.Host != s.HcloudCluster.Status.ControlPlaneLoadBalancer.IPv4 {
		p.KubeAPIServerDomain = &s.HcloudCluster.Spec.ControlPlaneEndpoint.Host
	} else {
		p.KubeAPIServerDomain = &emptyString
	}

	p.HcloudToken = &s.hcloudToken
	p.RobotUserName = &s.robotUserName
	p.RobotPassword = &s.robotPassword

	if s.HcloudCluster.Status.Network != nil {
		hcloudNetwork := intstr.FromInt(s.HcloudCluster.Status.Network.ID)
		p.HcloudNetwork = &hcloudNetwork
	}
	var port = strconv.FormatInt(int64(s.HcloudCluster.Spec.ControlPlaneEndpoint.Port), 10)
	p.Port = &port

	var CaCrtString string
	var CaKeyString string

	if s.HcloudCluster.Spec.VCKubeletClientSecretEnabled {
		secret := &corev1.Secret{}
		key := types.NamespacedName{Namespace: s.Namespace(), Name: s.Name() + "-ca"}
		if err := s.Client.Get(s.Ctx, key, secret); err != nil {
			return nil, errors.Wrapf(err, "failed to retrieve ca secret %s/%s", s.Namespace(), s.Name())
		}

		tlsCrt, ok := secret.Data["tls.crt"]

		if !ok {
			return nil, errors.New("error retrieving bootstrap data: secret value key is missing")
		}

		tlsKey, ok := secret.Data["tls.key"]
		if !ok {
			return nil, errors.New("error retrieving bootstrap data: secret value key is missing")
		}

		CaCrtString = string(tlsCrt)
		CaKeyString = string(tlsKey)
	}

	p.CAcrt = &CaCrtString
	p.CAkey = &CaKeyString
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

	_, err = clientSet.Discovery().RESTClient().Get().AbsPath("/readyz").DoRaw()
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
