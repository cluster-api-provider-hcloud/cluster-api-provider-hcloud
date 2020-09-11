package api

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// this variable needs to raised, to rebuild images (e.g. after packer config
// changes)
const imageVersion = 2

type HcloudClient interface {
	Token() string
	ListImages(context.Context, hcloud.ImageListOpts) ([]*hcloud.Image, error)
}

type PackerParameters struct {
	KubernetesVersion string
	Image             string
	// TODO add option to configure ContainerRuntime
	// TODO add option to configure OperatingSystem
}

func (p *PackerParameters) Hash() string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d", imageVersion)))
	h.Write([]byte(p.KubernetesVersion))
	h.Write([]byte(p.Image))
	return fmt.Sprintf("%x", h.Sum(nil))

}

func (p *PackerParameters) EnvironmentVariables() []string {
	return []string{
		fmt.Sprintf("PACKER_KUBERNETES_VERSION=%s", p.KubernetesVersion),
		fmt.Sprintf("PACKER_TEMPLATE_HASH=%s", p.Hash()),
		fmt.Sprintf("PACKER_IMAGE=%s", p.Image),
	}
}
