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
	// ClusterFinalizer allows Reconcile BareMetalMachine to clean up
	// resources associated with BareMetalMachine before removing it from the
	// apiserver.
	BareMetalMachineFinalizer = "baremetalmachine.cluster-api-provider-hcloud.capihc.com"
)

// BareMetalMachineSpec defines the desired state of a BareMetalMachine
type BareMetalMachineSpec struct {
	RobotTokenRef robotTokenRef `json:"robotTokenRef,omitempty"`

	SSHTokenRef sshTokenRef `json:"sshTokenRef,omitempty"`

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`

	// +optional
	Port *int `json:"port,omitempty"`

	// +optional
	Partition *string `json:"partition,omitempty"`

	ImagePath *string `json:"imagePath,omitempty"`
}

type robotTokenRef struct {
	PasswordKey string `json:"passwordKey,omitempty"`
	UserNameKey string `json:"userNameKey,omitempty"`
	TokenName   string `json:"tokenName,omitempty"`
}

type sshTokenRef struct {
	PublicKey  string `json:"publicKey,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	SSHKeyName string `json:"sshKeyName,omitempty"`
	TokenName  string `json:"tokenName,omitempty"`
}

//type BareMetalServerState string

// BareMetalMachineStatus defines the observed state of BareMetalMachine
type BareMetalMachineStatus struct {
	IPv4       string `json:"server_ip,omitempty"`
	IPv6       string `json:"ipv6,omitempty"`
	ServerID   int    `json:"server_number,omitempty"`
	ServerName string `json:"server_name,omitempty"`
	Status     string `json:"status,omitempty"`
	Cancelled  bool   `json:"cancelled,omitempty"`
	Reset      bool   `json:"reset,omitempty"`
	Rescue     bool   `json:"rescue,omitempty"`

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

func (h *BareMetalMachine) BareMetalMachineSpec() *BareMetalMachineSpec {
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

// BareMetalMachine is the Schema for the bareMetalMachine API
type BareMetalMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BareMetalMachineSpec   `json:"spec,omitempty"`
	Status BareMetalMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BareMetalMachineList contains a list of BareMetalMachine
type BareMetalMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BareMetalMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BareMetalMachine{}, &BareMetalMachineList{})
}
