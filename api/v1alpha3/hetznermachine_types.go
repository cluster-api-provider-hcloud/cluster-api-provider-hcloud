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
	// ClusterFinalizer allows ReconcileHetznerMachine to clean up Hetzner
	// resources associated with HetznerMachine before removing it from the
	// apiserver.
	MachineFinalizer = "hetznermachine.infrastructure.cluster.x-k8s.io"
)

// HetznerMachineSpec defines the desired state of HetznerMachine
type HetznerMachineSpec struct {
	Location HetznerLocation `json:"location,omitempty"`

	SSHKeys []HetznerSSHKeySpec `json:"sshKeys,omitempty"`

	Image *HetznerImageSpec `json:"image,omitempty"`

	Type HetznerMachineTypeSpec `json:"type,omitempty"`

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

type HetznerMachineTypeSpec string

type HetznerServerState string

type HetznerImageID int

type HetznerSSHKeySpec struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Name          *string               `json:"name,omitempty"`
	ID            *string               `json:"id,omitempty"`
}

type HetznerImageSpec struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
	Name          *string               `json:"name,omitempty"`
	ID            *HetznerImageID       `json:"id,omitempty"`
}

// HetznerMachineStatus defines the observed state of HetznerMachine
type HetznerMachineStatus struct {
	Location    HetznerLocation    `json:"location,omitempty"`
	NetworkZone HetznerNetworkZone `json:"networkZone,omitempty"`
	ImageID     *HetznerImageID    `json:"imageID,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the server's associated addresses.
	Addresses []v1.NodeAddress `json:"addresses,omitempty"`

	// ServerState is the state of the server for this machine.
	// +optional
	ServerState HetznerServerState `json:"serverState,omitempty"`

	// KubeadmConfigResourceVersionConfigured keeps track of the ResourceVersion which we last reconfigured KubeadmConfig
	// +optional
	KubeadmConfigResourceVersionUpdated *string `json:"kubeadmConfigResourceVersionUpdated"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hetznermachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.serverState",description="Server state"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="InstanceID",type="string",JSONPath=".spec.providerID",description="Hetzner instance ID"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this HetznerMachine"

// HetznerMachine is the Schema for the hetznermachine API
type HetznerMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HetznerMachineSpec   `json:"spec,omitempty"`
	Status HetznerMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HetznerMachineList contains a list of HetznerMachine
type HetznerMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HetznerMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HetznerMachine{}, &HetznerMachineList{})
}
