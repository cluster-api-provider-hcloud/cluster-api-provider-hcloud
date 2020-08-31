package scope

import (
	"context"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// HcloudClient collects all methods used by the controller in the hcloud cloud API
type HcloudClient interface {
	Token() string
	ListLocation(context.Context) ([]*hcloud.Location, error)
	CreateLoadBalancer(context.Context, hcloud.LoadBalancerCreateOpts) (hcloud.LoadBalancerCreateResult, *hcloud.Response, error)
	DeleteLoadBalancer(context.Context, *hcloud.LoadBalancer) (*hcloud.Response, error)
	ListLoadBalancers(context.Context, hcloud.LoadBalancerListOpts) ([]*hcloud.LoadBalancer, error)
	GetLoadBalancerTypeByName(context.Context, string) (*hcloud.LoadBalancerType, *hcloud.Response, error)
	AddTargetServerToLoadBalancer(context.Context, hcloud.LoadBalancerAddServerTargetOpts, *hcloud.LoadBalancer) (*hcloud.Action, *hcloud.Response, error)
	DeleteTargetServerOfLoadBalancer(context.Context, *hcloud.LoadBalancer, *hcloud.Server) (*hcloud.Action, *hcloud.Response, error)
	ListImages(context.Context, hcloud.ImageListOpts) ([]*hcloud.Image, error)
	CreateServer(context.Context, hcloud.ServerCreateOpts) (hcloud.ServerCreateResult, *hcloud.Response, error)
	ListServers(context.Context, hcloud.ServerListOpts) ([]*hcloud.Server, error)
	DeleteServer(context.Context, *hcloud.Server) (*hcloud.Response, error)
	ShutdownServer(context.Context, *hcloud.Server) (*hcloud.Action, *hcloud.Response, error)
	CreateVolume(context.Context, hcloud.VolumeCreateOpts) (hcloud.VolumeCreateResult, *hcloud.Response, error)
	ListVolumes(context.Context, hcloud.VolumeListOpts) ([]*hcloud.Volume, error)
	DeleteVolume(context.Context, *hcloud.Volume) (*hcloud.Response, error)
	CreateNetwork(context.Context, hcloud.NetworkCreateOpts) (*hcloud.Network, *hcloud.Response, error)
	ListNetworks(context.Context, hcloud.NetworkListOpts) ([]*hcloud.Network, error)
	DeleteNetwork(context.Context, *hcloud.Network) (*hcloud.Response, error)
	ListSSHKeys(ctx context.Context, opts hcloud.SSHKeyListOpts) ([]*hcloud.SSHKey, *hcloud.Response, error)
}

type HcloudClientFactory func(context.Context) (HcloudClient, error)

var _ HcloudClient = &realClient{}

type realClient struct {
	client *hcloud.Client
	token  string
}

func (c *realClient) Token() string {
	return c.token
}

func (c *realClient) ListLocation(ctx context.Context) ([]*hcloud.Location, error) {
	return c.client.Location.All(ctx)
}

func (c *realClient) CreateLoadBalancer(ctx context.Context, opts hcloud.LoadBalancerCreateOpts) (hcloud.LoadBalancerCreateResult, *hcloud.Response, error) {
	return c.client.LoadBalancer.Create(ctx, opts)
}

func (c *realClient) DeleteLoadBalancer(ctx context.Context, loadBalancer *hcloud.LoadBalancer) (*hcloud.Response, error) {
	return c.client.LoadBalancer.Delete(ctx, loadBalancer)
}

func (c *realClient) ListLoadBalancers(ctx context.Context, opts hcloud.LoadBalancerListOpts) ([]*hcloud.LoadBalancer, error) {
	return c.client.LoadBalancer.AllWithOpts(ctx, opts)
}

func (c *realClient) GetLoadBalancerTypeByName(ctx context.Context, name string) (*hcloud.LoadBalancerType, *hcloud.Response, error) {
	return c.client.LoadBalancerType.GetByName(ctx, name)
}

func (c *realClient) AddTargetServerToLoadBalancer(ctx context.Context, opts hcloud.LoadBalancerAddServerTargetOpts, lb *hcloud.LoadBalancer) (*hcloud.Action, *hcloud.Response, error) {
	return c.client.LoadBalancer.AddServerTarget(ctx, lb, opts)
}

func (c *realClient) DeleteTargetServerOfLoadBalancer(ctx context.Context, lb *hcloud.LoadBalancer, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error) {
	return c.client.LoadBalancer.RemoveServerTarget(ctx, lb, server)
}

func (c *realClient) ListImages(ctx context.Context, opts hcloud.ImageListOpts) ([]*hcloud.Image, error) {
	return c.client.Image.AllWithOpts(ctx, opts)
}

func (c *realClient) CreateServer(ctx context.Context, opts hcloud.ServerCreateOpts) (hcloud.ServerCreateResult, *hcloud.Response, error) {
	return c.client.Server.Create(ctx, opts)
}

func (c *realClient) ListServers(ctx context.Context, opts hcloud.ServerListOpts) ([]*hcloud.Server, error) {
	return c.client.Server.AllWithOpts(ctx, opts)
}

func (c *realClient) ShutdownServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error) {

	return c.client.Server.Shutdown(ctx, server)
}

func (c *realClient) DeleteServer(ctx context.Context, server *hcloud.Server) (*hcloud.Response, error) {
	return c.client.Server.Delete(ctx, server)
}

func (c *realClient) CreateVolume(ctx context.Context, opts hcloud.VolumeCreateOpts) (hcloud.VolumeCreateResult, *hcloud.Response, error) {
	return c.client.Volume.Create(ctx, opts)
}

func (c *realClient) ListVolumes(ctx context.Context, opts hcloud.VolumeListOpts) ([]*hcloud.Volume, error) {
	return c.client.Volume.AllWithOpts(ctx, opts)
}

func (c *realClient) DeleteVolume(ctx context.Context, server *hcloud.Volume) (*hcloud.Response, error) {
	return c.client.Volume.Delete(ctx, server)
}

func (c *realClient) CreateNetwork(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, *hcloud.Response, error) {
	return c.client.Network.Create(ctx, opts)
}

func (c *realClient) ListNetworks(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error) {
	return c.client.Network.AllWithOpts(ctx, opts)
}

func (c *realClient) DeleteNetwork(ctx context.Context, server *hcloud.Network) (*hcloud.Response, error) {
	return c.client.Network.Delete(ctx, server)
}

func (c *realClient) ListSSHKeys(ctx context.Context, opts hcloud.SSHKeyListOpts) ([]*hcloud.SSHKey, *hcloud.Response, error) {
	return c.client.SSHKey.List(ctx, opts)
}
