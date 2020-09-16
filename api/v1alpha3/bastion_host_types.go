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
)

type BastionHostSpec struct {
	Name       *string            `json:"name,omitempty"`
	ServerType *string            `json:"serverType,omitempty"`
	Location   *string            `json:"location,omitempty"`
	SSHKeys    []HcloudSSHKeySpec `json:"sshKeys,omitempty"`
	ImageName  string             `json:"imageName,omitempty"`
	Version    *string            `json:"version,omitempty"`
}

type BastionHostStatus struct {
	Location string         `json:"location,omitempty"`
	ImageID  *HcloudImageID `json:"imageID,omitempty"`
}

func (h *BastionHost) BastionHostSpec() *BastionHostSpec {
	return h.Spec.DeepCopy()
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=bastionhosts,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// BastionHost is the Schema for the bastion host API
type BastionHost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BastionHostSpec   `json:"spec,omitempty"`
	Status BastionHostStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BastionHostList contains a list of BastionHosts
type BastionHostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BastionHost `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BastionHost{}, &BastionHostList{})
}
