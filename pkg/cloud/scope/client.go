package scope

import (
	"context"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// HetznerClient collects all methods used by the controller in the hetzner cloud API
type HetznerClient interface {
	ListLocation(context.Context) ([]*hcloud.Location, error)
	CreateFloatingIP(context.Context, hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error)
	DeleteFloatingIP(context.Context, *hcloud.FloatingIP) (*hcloud.Response, error)
	ListFloatingIPs(context.Context, hcloud.FloatingIPListOpts) ([]*hcloud.FloatingIP, error)
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
}

type HetznerClientFactory func(context.Context) (HetznerClient, error)

var _ HetznerClient = &realClient{}

type realClient struct {
	client *hcloud.Client
}

func (c *realClient) ListLocation(ctx context.Context) ([]*hcloud.Location, error) {
	return c.client.Location.All(ctx)
}

func (c *realClient) CreateFloatingIP(ctx context.Context, opts hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error) {
	return c.client.FloatingIP.Create(ctx, opts)
}

func (c *realClient) DeleteFloatingIP(ctx context.Context, ip *hcloud.FloatingIP) (*hcloud.Response, error) {
	return c.client.FloatingIP.Delete(ctx, ip)
}

func (c *realClient) ListFloatingIPs(ctx context.Context, opts hcloud.FloatingIPListOpts) ([]*hcloud.FloatingIP, error) {
	return c.client.FloatingIP.AllWithOpts(ctx, opts)
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
