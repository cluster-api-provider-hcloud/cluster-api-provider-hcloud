package api

import (
	"crypto/sha256"
	"fmt"
)

// this variable needs to raised, to rebuild images (e.g. after packer config
// changes)
const imageVersion = 1

type PackerParameters struct {
	KubernetesVersion string
	// TODO add option to configure ContainerRuntime
	// TODO add option to configure OperatingSystem
}

func (p *PackerParameters) Hash() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d", imageVersion)))
	h.Write([]byte(p.KubernetesVersion))
	return fmt.Sprintf("%x", h.Sum(nil))

}

func (p *PackerParameters) EnvironmentVariables() []string {
	return []string{
		fmt.Sprintf("PACKER_KUBERNETES_VERSION=%s", p.KubernetesVersion),
		fmt.Sprintf("PACKER_TEMPLATE_HASH=%s", p.Hash()),
	}
}
