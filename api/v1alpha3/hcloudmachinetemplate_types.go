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

// HcloudMachineTemplateSpec defines the desired state of HcloudMachineTemplate
type HcloudMachineTemplateSpec struct {
	Template HcloudMachineTemplateResource `json:"template"`
}

// HcloudMachineTemplateResource describes the data needed to create am HcloudMachine from a template
type HcloudMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec HcloudMachineSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hcloudmachinetemplates,scope=Namespaced,categories=cluster-api

// HcloudMachineTemplate is the Schema for the hcloudmachine API
type HcloudMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HcloudMachineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// HcloudMachineTemplateList contains a list of HcloudMachineTemplate
type HcloudMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HcloudMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HcloudMachineTemplate{}, &HcloudMachineTemplateList{})
}
