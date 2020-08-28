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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// VolumeFinalizer allows ReconcileHcloudVolume to clean up Hcloud
	// resources associated with HcloudVolume before removing it from the
	// apiserver.
	VolumeFinalizer = "hcloudvolume.cluster-api-provider-hcloud.capihc.com"
)

type HcloudVolumeID int

// HcloudVolumeReclaimPolicy describes a policy for end-of-life maintenance of persistent volumes
type HcloudVolumeReclaimPolicy string

const (
	// HcloudVolumeReclaimDelete means the volume will be deleted from
	// Kubernetes on release from its claim.  The volume plugin must support
	// Deletion.
	HcloudVolumeReclaimDelete HcloudVolumeReclaimPolicy = "Delete"
	// HcloudVolumeReclaimRetain means the volume will be left in its current
	// phase (Released) for manual reclamation by the administrator.  The
	// default policy is Retain.
	HcloudVolumeReclaimRetain HcloudVolumeReclaimPolicy = "Retain"
)

// HcloudVolumeSpec defines the desired state of HcloudVolume
type HcloudVolumeSpec struct {
	Location HcloudLocation `json:"location,omitempty"`

	// Size contains the minimum requested size of the volume
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// Size contains the minimum requested size of the volume
	// +optional
	ReclaimPolicy HcloudVolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
}

// HcloudVolumeStatus defines the observed state of HcloudVolume
type HcloudVolumeStatus struct {
	Location HcloudLocation `json:"location,omitempty"`

	// Size contains the actual size of the volume
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// VolumeID contains the ID of the releated volume
	VolumeID *HcloudVolumeID `json:"volumeID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hcloudvolumes,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".status.location",description="Location of the volume"
// +kubebuilder:subresource:status

// HcloudVolume is the Schema for the hcloudvolumes API
type HcloudVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HcloudVolumeSpec   `json:"spec,omitempty"`
	Status HcloudVolumeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HcloudVolumeList contains a list of HcloudVolume
type HcloudVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HcloudVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HcloudVolume{}, &HcloudVolumeList{})
}
