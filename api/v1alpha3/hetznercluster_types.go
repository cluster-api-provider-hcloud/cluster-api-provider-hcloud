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

	SSHKeys []HetznerSSHKeySpec `json:"sshKeys,omitempty"`

	Image *HetznerImageSpec `json:"image,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	ControlPlaneFloatingIPs []HetznerFloatingIPSpec `json:"controlPlaneFloatingIPs,omitempty"`

	TokenRef *corev1.SecretKeySelector `json:"tokenRef,omitempty"`
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
	Type    HetznerFloatingIPType `json:"type"`
}

type HetznerSSHKeySpec struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Name          *string               `json:"name,omitempty"`
	ID            *string               `json:"id,omitempty"`
}

type HetznerImageSpec struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Name          *string               `json:"name,omitempty"`
	ID            *string               `json:"id,omitempty"`
}

// HetznerClusterStatus defines the observed state of HetznerCluster
type HetznerClusterStatus struct {
	Location                HetznerLocation           `json:"location,omitempty"`
	NetworkZone             HetznerNetworkZone        `json:"networkZone,omitempty"`
	ControlPlaneFloatingIPs []HetznerFloatingIPStatus `json:"controlPlaneFloatingIPs,omitempty"`

	ImageID string `json:"imageID,omitempty"`
}

// +kubebuilder:object:root=true
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
