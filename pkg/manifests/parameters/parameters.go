package parameters

import (
	"fmt"
	"net"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"
)

type ManifestParameters struct {
	HcloudToken         *string
	HcloudNetwork       *intstr.IntOrString
	KubeAPIServerIPv4   *string
	KubeAPIServerDomain *string
	PodCIDRBlock        *net.IPNet
	Manifests           []string
}

func (m *ManifestParameters) ExtVar() map[string]string {
	extVar := make(map[string]string)

	if key, val := "kube-apiserver-ip", m.KubeAPIServerIPv4; val != nil {
		extVar[key] = *val
	} else {
		extVar[key] = ""
	}

	fmt.Printf("Parameters.go: %s", *m.KubeAPIServerDomain)

	if key, val := "kube-apiserver-domain", m.KubeAPIServerDomain; val != nil {
		extVar[key] = *val
	} else {
		extVar[key] = ""
	}

	extVar["manifests"] = strings.Join(m.Manifests, ",")

	if key, val := "hcloud-token", m.HcloudToken; val != nil {
		extVar[key] = *val
	}

	if key, val := "hcloud-network", m.HcloudNetwork; val != nil {
		extVar[key] = val.String()
	} else {
		extVar[key] = ""
	}

	if key, val := "pod-cidr-block", m.PodCIDRBlock; val != nil {
		extVar[key] = val.String()
	}

	return extVar
}
