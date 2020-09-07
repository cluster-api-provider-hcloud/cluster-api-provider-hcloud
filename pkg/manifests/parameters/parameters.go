package parameters

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ManifestParameters struct {
	HcloudToken             *string
	HcloudNetwork           *intstr.IntOrString
	HcloudLoadBalancerIPv4s []string
	PodCIDRBlock            *net.IPNet
	NewManifests            []string

	Network *ManifestNetwork
}

type ManifestNetwork struct {
	Calico  *infrav1.HcloudClusterSpecManifestsNetworkCalico  `json:"calico,omitempty"`
	Cilium  *ManifestNetworkCilium                            `json:"cilium,omitempty"`
	Flannel *infrav1.HcloudClusterSpecManifestsNetworkFlannel `json:"flannel,omitempty"`
}

type ManifestNetworkCilium struct {
	IPSecKeys *string `json:"ipSecKeys,omitempty"`
}

func (m *ManifestParameters) ExtVar() map[string]string {
	extVar := make(map[string]string)

	extVar["hcloud-loadbalancer"] = strings.Join(m.HcloudLoadBalancerIPv4s, ",")

	extVar["manifests"] = strings.Join(m.NewManifests, ",")

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

	if key, val := "network", m.Network; val != nil {
		data, err := json.Marshal(val)
		if err != nil {
			panic(fmt.Sprintf("unexpected error marshaling network JSON: %v", err))
		}
		extVar[key] = string(data)
	} else {
		extVar[key] = "{}"
	}
	return extVar
}
