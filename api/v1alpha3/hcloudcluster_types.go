/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

const (
	// ClusterFinalizer allows ReconcileHcloudCluster to clean up Hcloud
	// resources associated with HcloudCluster before removing it from the
	// apiserver.
	ClusterFinalizer = "hcloudcluster.cluster-api-provider-hcloud.swine.dev"
)

type HcloudLocation string
type HcloudNetworkZone string

// +kubebuilder:validation:Enum=IPv4;IPv6
type HcloudFloatingIPType string

const (
	HcloudFloatingIPTypeIPv4 = HcloudFloatingIPType("IPv4")
	HcloudFloatingIPTypeIPv6 = HcloudFloatingIPType("IPv6")
)

// HcloudClusterSpec defines the desired state of HcloudCluster
type HcloudClusterSpec struct {
	Location HcloudLocation `json:"location,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	ControlPlaneFloatingIPs []HcloudFloatingIPSpec `json:"controlPlaneFloatingIPs,omitempty"`

	Network *HcloudNetworkSpec `json:"network,omitempty"`

	TokenRef *corev1.SecretKeySelector `json:"tokenRef,omitempty"`
}

type HcloudNetwork struct {
	CIDRBlock string `json:"cidrBlock,omitempty"`
}

type HcloudNetworkSpec struct {
	HcloudNetwork `json:",inline"`

	Subnets []HcloudNetworkSubnetSpec `json:"subnets,omitempty"`
}

type HcloudNetworkSubnetSpec struct {
	HcloudNetwork `json:",inline"`

	NetworkZone HcloudNetworkZone `json:"networkZone,omitempty"`
}

type HcloudNetworkStatus struct {
	HcloudNetworkSpec `json:",inline"`

	ID     int               `json:"id,omitempty"`
	Labels map[string]string `json:"-"`
}

type HcloudFloatingIPSpec struct {
	Name *string              `json:"name,omitempty"`
	ID   *int                 `json:"id,omitempty"`
	Type HcloudFloatingIPType `json:"type"`
}

type HcloudFloatingIPStatus struct {
	ID      int                  `json:"id,omitempty"`
	Name    string               `json:"name,omitempty"`
	Network string               `json:"network,omitempty"`
	IP      string               `json:"ip,omitempty"`
	Type    HcloudFloatingIPType `json:"type"`
	Labels  map[string]string    `json:"-"`
}

// HcloudClusterStatus defines the observed state of HcloudCluster
type HcloudClusterStatus struct {
	Location                HcloudLocation           `json:"location,omitempty"`
	NetworkZone             HcloudNetworkZone        `json:"networkZone,omitempty"`
	ControlPlaneFloatingIPs []HcloudFloatingIPStatus `json:"controlPlaneFloatingIPs,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// +optional
	Network *HcloudNetworkStatus `json:"network,omitempty"`

	// APIEndpoints represents the endpoints to communicate with the control plane.
	// +optional
	APIEndpoints []clusterv1.APIEndpoint `json:"apiEndpoints,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hcloudclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".status.location",description="Location of the cluster"
// +kubebuilder:printcolumn:name="NetworkZone",type="string",JSONPath=".status.networkZone",description="NetworkZone of the cluster"
// +kubebuilder:subresource:status

// HcloudCluster is the Schema for the hcloudclusters API
type HcloudCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HcloudClusterSpec   `json:"spec,omitempty"`
	Status HcloudClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HcloudClusterList contains a list of HcloudCluster
type HcloudClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HcloudCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HcloudCluster{}, &HcloudClusterList{})
}
