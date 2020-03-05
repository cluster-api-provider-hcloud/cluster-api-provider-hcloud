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
	// VolumeFinalizer allows ReconcileHetznerVolume to clean up Hetzner
	// resources associated with HetznerVolume before removing it from the
	// apiserver.
	VolumeFinalizer = "hetznervolume.infrastructure.cluster.x-k8s.io"
)

type HetznerVolumeID int

// HetznerVolumeReclaimPolicy describes a policy for end-of-life maintenance of persistent volumes
type HetznerVolumeReclaimPolicy string

const (
	// HetznerVolumeReclaimDelete means the volume will be deleted from
	// Kubernetes on release from its claim.  The volume plugin must support
	// Deletion.
	HetznerVolumeReclaimDelete HetznerVolumeReclaimPolicy = "Delete"
	// HetznerVolumeReclaimRetain means the volume will be left in its current
	// phase (Released) for manual reclamation by the administrator.  The
	// default policy is Retain.
	HetznerVolumeReclaimRetain HetznerVolumeReclaimPolicy = "Retain"
)

// HetznerVolumeSpec defines the desired state of HetznerVolume
type HetznerVolumeSpec struct {
	Location HetznerLocation `json:"location,omitempty"`

	// Size contains the minimum requested size of the volume
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// Size contains the minimum requested size of the volume
	// +optional
	ReclaimPolicy HetznerVolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
}

// HetznerVolumeStatus defines the observed state of HetznerVolume
type HetznerVolumeStatus struct {
	Location HetznerLocation `json:"location,omitempty"`

	// Size contains the actual size of the volume
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// VolumeID contains the ID of the releated volume
	VolumeID *HetznerVolumeID `json:"volumeID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hetznervolumes,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".status.location",description="Location of the volume"
// +kubebuilder:subresource:status

// HetznerVolume is the Schema for the hetznervolumes API
type HetznerVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HetznerVolumeSpec   `json:"spec,omitempty"`
	Status HetznerVolumeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HetznerVolumeList contains a list of HetznerVolume
type HetznerVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HetznerVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HetznerVolume{}, &HetznerVolumeList{})
}
