package parameters

import (
	"net"

	"k8s.io/apimachinery/pkg/util/intstr"
)

type ManifestParameters struct {
	HcloudToken   *string
	HcloudNetwork *intstr.IntOrString
	PodCIDRBlock  *net.IPNet
}

func (m *ManifestParameters) ExtVar() map[string]string {
	extVar := make(map[string]string)

	if key, val := "hcloud-token", m.HcloudToken; val != nil {
		extVar[key] = *val
	}

	if key, val := "hcloud-network", m.HcloudNetwork; val != nil {
		extVar[key] = val.String()
	}

	if key, val := "pod-cidr-block", m.PodCIDRBlock; val != nil {
		extVar[key] = val.String()
	}

	return extVar
}
