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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// ClusterFinalizer allows ReconcileHcloudCluster to clean up Hcloud
	// resources associated with HcloudCluster before removing it from the
	// apiserver.
	ClusterFinalizer = "hcloudcluster.cluster-api-provider-hcloud.capihc.com"
)

type HcloudLocation string
type HcloudNetworkZone string

// +kubebuilder:validation:Enum=round_robin;least_connections
type HcloudLoadBalancerAlgorithmType string

const (
	HcloudLoadBalancerAlgorithmTypeRoundRobin       = HcloudLoadBalancerAlgorithmType("round_robin")
	HcloudLoadBalancerAlgorithmTypeLeastConnections = HcloudLoadBalancerAlgorithmType("least_connections")
)

// HcloudClusterSpec defines the desired state of HcloudCluster
type HcloudClusterSpec struct {
	Locations []HcloudLocation `json:"locations,omitempty"`

	// define cluster wide SSH keys
	SSHKeys []HcloudSSHKeySpec `json:"sshKeys,omitempty"`

	ControlPlaneLoadBalancer HcloudLoadBalancerSpec `json:"controlPlaneLoadbalancer,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint *clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`

	Network *HcloudNetworkSpec `json:"network,omitempty"`

	TokenRef *corev1.SecretKeySelector `json:"tokenRef,omitempty"`
}

type HcloudNetwork struct {
	CIDRBlock string `json:"cidrBlock,omitempty"`
}

type HcloudNetworkSpec struct {
	HcloudNetwork `json:",inline"`

	Subnets []HcloudNetworkSubnetSpec `json:"subnets,omitempty"`
}

func (s *HcloudNetworkSpec) IsZero() bool {
	if len(s.CIDRBlock) > 0 {
		return false
	}
	if len(s.Subnets) > 0 {
		return false
	}
	return true
}

type HcloudNetworkSubnetSpec struct {
	HcloudNetwork `json:",inline"`

	NetworkZone HcloudNetworkZone `json:"networkZone,omitempty"`
}

type HcloudNetworkStatus struct {
	HcloudNetworkSpec `json:",inline"`

	ID     int               `json:"id,omitempty"`
	Labels map[string]string `json:"-"`
}

type HcloudLoadBalancerSpec struct {
	Name      *string                         `json:"name,omitempty"`
	ID        *int                            `json:"id,omitempty"`
	Algorithm HcloudLoadBalancerAlgorithmType `json:"algorithm,omitempty"`
	Type      string                          `json:"type,omitempty"`
}
type HcloudLoadBalancerStatus struct {
	ID         int                             `json:"id,omitempty"`
	Name       string                          `json:"name,omitempty"`
	Type       string                          `json:"type,omitempty"`
	IPv4       string                          `json:"ipv4,omitempty"`
	IPv6       string                          `json:"ipv6,omitempty"`
	InternalIP string                          `json:"internalIP,omitempty"`
	Labels     map[string]string               `json:"-"`
	Algorithm  HcloudLoadBalancerAlgorithmType `json:"algorithm,omitempty"`
	Targets    []int                           `json:"-"`
	HasNetwork bool                            `json:"hasNetwork"`
}

type HcloudClusterStatusManifests struct {
	Initialized *bool   `json:"initialized,omitempty"`
	AppliedHash *string `json:"appliedHash,omitempty"`
}

// HcloudClusterStatus defines the observed state of HcloudCluster
type HcloudClusterStatus struct {
	Locations                []HcloudLocation         `json:"locations,omitempty"`
	NetworkZone              HcloudNetworkZone        `json:"networkZone,omitempty"`
	ControlPlaneLoadBalancer HcloudLoadBalancerStatus `json:"controlPlaneLoadBalancer,omitempty"`

	// +optional
	Network *HcloudNetworkStatus `json:"network,omitempty"`

	// Manifests stores the if the cluster has already applied the minimal
	// manifests
	// +optional
	Manifests *HcloudClusterStatusManifests `json:"manifests,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// FailureDomains contains the failure domains that machines should be
	// placed in. failureDomains is a map, defined as
	// map[string]FailureDomainSpec. A unique key must be used for each
	// FailureDomainSpec
	FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hcloudclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Locations",type="string",JSONPath=".status.locations",description="Locations of the cluster"
// +kubebuilder:printcolumn:name="NetworkZone",type="string",JSONPath=".status.networkZone",description="NetworkZone of the cluster"
// +kubebuilder:subresource:status

// HcloudCluster is the Schema for the hcloudclusters API
type HcloudCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HcloudClusterSpec   `json:"spec,omitempty"`
	Status HcloudClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HcloudClusterList contains a list of HcloudCluster
type HcloudClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HcloudCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HcloudCluster{}, &HcloudClusterList{})
}
