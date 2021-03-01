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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// ClusterFinalizer allows Reconcile HetznerBareMetalMachine to clean up
	// resources associated with HetznerBareMetalMachine before removing it from the
	// apiserver.
	HetznerBareMetalMachineFinalizer = "hetznerbaremetalmachine.cluster-api-provider-hcloud.capihc.com"
)

// BareMetalMachineSpec defines the desired state of a BareMetalMachine
type HetznerBareMetalMachineSpec struct {
	SSHTokenRef sshTokenRef `json:"sshTokenRef"`

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID"`

	// +optional
	Partition *string `json:"partition"`

	ImagePath *string `json:"imagePath"`

	ServerType *string `json:"serverType"`
}

type sshTokenRef struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
	SSHKeyName string `json:"sshKeyName"`
	TokenName  string `json:"tokenName"`
}

//type HetznerBareMetalServerState string

// HetznerBareMetalMachineStatus defines the observed state of HetznerBareMetalMachine
type HetznerBareMetalMachineStatus struct {
	IPv4        string `json:"server_ip,omitempty"`
	IPv6        string `json:"ipv6,omitempty"`
	ServerID    int    `json:"server_number,omitempty"`
	ServerName  string `json:"server_name,omitempty"`
	Ready       bool   `json:"ready,omitempty"`
	ServerState string `json:"serverState,omitempty"`
	Cancelled   bool   `json:"cancelled,omitempty"`
	Reset       bool   `json:"reset,omitempty"`
	Rescue      bool   `json:"rescue,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *errors.MachineStatusError `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

func (h *HetznerBareMetalMachine) HetznerBareMetalMachineSpec() *HetznerBareMetalMachineSpec {
	return h.Spec.DeepCopy()
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=baremetalmachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.serverState",description="Server state"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="InstanceID",type="string",JSONPath=".spec.providerID",description="Hcloud instance ID"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this BareMetalMachine"

// HetznerBareMetalMachine is the Schema for the bareMetalMachine API
type HetznerBareMetalMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HetznerBareMetalMachineSpec   `json:"spec,omitempty"`
	Status HetznerBareMetalMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HetznerBareMetalMachineList contains a list of BareMetalMachine
type HetznerBareMetalMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HetznerBareMetalMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HetznerBareMetalMachine{}, &HetznerBareMetalMachineList{})
}
