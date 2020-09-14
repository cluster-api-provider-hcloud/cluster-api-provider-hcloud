package parameters

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"
)

type ManifestParameters struct {
	HcloudToken             *string
	HcloudNetwork           *intstr.IntOrString
	HcloudLoadBalancerIPv4s []string
}

func (m *ManifestParameters) ExtVar() map[string]string {
	extVar := make(map[string]string)

	extVar["hcloud-loadbalancer"] = strings.Join(m.HcloudLoadBalancerIPv4s, ",")

	if key, val := "hcloud-token", m.HcloudToken; val != nil {
		extVar[key] = *val
	}

	if key, val := "hcloud-network", m.HcloudNetwork; val != nil {
		extVar[key] = val.String()
	} else {
		extVar[key] = ""
	}

	return extVar
}
