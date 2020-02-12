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
	// ClusterFinalizer allows ReconcileHetznerCluster to clean up Hetzner
	// resources associated with HetznerCluster before removing it from the
	// apiserver.
	ClusterFinalizer = "hetznercluster.infrastructure.cluster.x-k8s.io"
)

type HetznerLocation string
type HetznerNetworkZone string

// +kubebuilder:validation:Enum=IPv4;IPv6
type HetznerFloatingIPType string

const (
	HetznerFloatingIPTypeIPv4 = HetznerFloatingIPType("IPv4")
	HetznerFloatingIPTypeIPv6 = HetznerFloatingIPType("IPv6")
)

// HetznerClusterSpec defines the desired state of HetznerCluster
type HetznerClusterSpec struct {
	Location HetznerLocation `json:"location,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	ControlPlaneFloatingIPs []HetznerFloatingIPSpec `json:"controlPlaneFloatingIPs,omitempty"`

	Network *HetznerNetworkSpec `json:"network,omitempty"`

	TokenRef *corev1.SecretKeySelector `json:"tokenRef,omitempty"`
}

type HetznerNetwork struct {
	CIDRBlock string `json:"cidrBlock,omitempty"`
}

type HetznerNetworkSpec struct {
	HetznerNetwork `json:",inline"`

	Subnets []HetznerNetworkSubnetSpec `json:"subnets,omitempty"`
}

type HetznerNetworkSubnetSpec struct {
	HetznerNetwork `json:",inline"`

	NetworkZone HetznerNetworkZone `json:"networkZone,omitempty"`
}

type HetznerNetworkStatus struct {
	HetznerNetworkSpec `json:",inline"`

	ID     int               `json:"id,omitempty"`
	Labels map[string]string `json:"-"`
}

type HetznerFloatingIPSpec struct {
	Name *string               `json:"name,omitempty"`
	ID   *int                  `json:"id,omitempty"`
	Type HetznerFloatingIPType `json:"type"`
}

type HetznerFloatingIPStatus struct {
	ID      int                   `json:"id,omitempty"`
	Name    string                `json:"name,omitempty"`
	Network string                `json:"network,omitempty"`
	IP      string                `json:"ip,omitempty"`
	Type    HetznerFloatingIPType `json:"type"`
	Labels  map[string]string     `json:"-"`
}

// HetznerClusterStatus defines the observed state of HetznerCluster
type HetznerClusterStatus struct {
	Location                HetznerLocation           `json:"location,omitempty"`
	NetworkZone             HetznerNetworkZone        `json:"networkZone,omitempty"`
	ControlPlaneFloatingIPs []HetznerFloatingIPStatus `json:"controlPlaneFloatingIPs,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// +optional
	Network *HetznerNetworkStatus `json:"network,omitempty"`

	// APIEndpoints represents the endpoints to communicate with the control plane.
	// +optional
	APIEndpoints []clusterv1.APIEndpoint `json:"apiEndpoints,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hetznerclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".status.location",description="Location of the cluster"
// +kubebuilder:printcolumn:name="NetworkZone",type="string",JSONPath=".status.networkZone",description="NetworkZone of the cluster"
// +kubebuilder:subresource:status

// HetznerCluster is the Schema for the hetznerclusters API
type HetznerCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HetznerClusterSpec   `json:"spec,omitempty"`
	Status HetznerClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HetznerClusterList contains a list of HetznerCluster
type HetznerClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HetznerCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HetznerCluster{}, &HetznerClusterList{})
}
