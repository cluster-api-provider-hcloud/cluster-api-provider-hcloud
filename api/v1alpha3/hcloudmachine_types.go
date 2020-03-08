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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ClusterFinalizer allows ReconcileHcloudMachine to clean up Hcloud
	// resources associated with HcloudMachine before removing it from the
	// apiserver.
	MachineFinalizer = "hcloudmachine.cluster-api-provider-hcloud.swine.dev"
)

// HcloudMachineSpec defines the desired state of HcloudMachine
type HcloudMachineSpec struct {
	Location HcloudLocation `json:"location,omitempty"`

	SSHKeys []HcloudSSHKeySpec `json:"sshKeys,omitempty"`

	Image *HcloudImageSpec `json:"image,omitempty"`

	Type HcloudMachineTypeSpec `json:"type,omitempty"`

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

type HcloudMachineTypeSpec string

type HcloudServerState string

type HcloudImageID int

type HcloudSSHKeySpec struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Name          *string               `json:"name,omitempty"`
	ID            *string               `json:"id,omitempty"`
}

type HcloudImageSpec struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Name          *string               `json:"name,omitempty"`
	ID            *HcloudImageID        `json:"id,omitempty"`
}

// HcloudMachineStatus defines the observed state of HcloudMachine
type HcloudMachineStatus struct {
	Location    HcloudLocation    `json:"location,omitempty"`
	NetworkZone HcloudNetworkZone `json:"networkZone,omitempty"`
	ImageID     *HcloudImageID    `json:"imageID,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the server's associated addresses.
	Addresses []v1.NodeAddress `json:"addresses,omitempty"`

	// ServerState is the state of the server for this machine.
	// +optional
	ServerState HcloudServerState `json:"serverState,omitempty"`

	// KubeadmConfigResourceVersionConfigured keeps track of the ResourceVersion which we last reconfigured KubeadmConfig
	// +optional
	KubeadmConfigResourceVersionUpdated *string `json:"kubeadmConfigResourceVersionUpdated"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hcloudmachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.serverState",description="Server state"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="InstanceID",type="string",JSONPath=".spec.providerID",description="Hcloud instance ID"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this HcloudMachine"

// HcloudMachine is the Schema for the hcloudmachine API
type HcloudMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HcloudMachineSpec   `json:"spec,omitempty"`
	Status HcloudMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HcloudMachineList contains a list of HcloudMachine
type HcloudMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HcloudMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HcloudMachine{}, &HcloudMachineList{})
}
