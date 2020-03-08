/*
Copyright 2019 The Kubernetes Authors.

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

import "fmt"

// ResourceLifecycle configures the lifecycle of a resource
type ResourceLifecycle string

const (
	// ResourceLifecycleOwned is the value we use when tagging resources to indicate
	// that the resource is considered owned and managed by the cluster,
	// and in particular that the lifecycle is tied to the lifecycle of the cluster.
	ResourceLifecycleOwned = ResourceLifecycle("owned")

	// ResourceLifecycleShared is the value we use when tagging resources to indicate
	// that the resource is shared between multiple clusters, and should not be destroyed
	// if the cluster is destroyed.
	ResourceLifecycleShared = ResourceLifecycle("shared")

	// NameKubernetesHcloudCloudProviderPrefix is the tag name used by the cloud provider to logically
	// separate independent cluster resources. We use it to identify which resources we expect
	// to be permissive about state changes.
	// logically independent clusters running in the same AZ.
	// The tag key = NameKubernetesHcloudCloudProviderPrefix + clusterID
	// The tag value is an ownership value
	NameKubernetesHcloudCloudProviderPrefix = "kubernetes.io/cluster/"

	// NameHcloudProviderPrefix is the tag prefix we use to differentiate
	// cluster-api-provider-hcloud owned components from other tooling that
	// uses NameKubernetesClusterPrefix
	NameHcloudProviderPrefix = "cluster-api-provider-hcloud.sigs.k8s.io/"

	// NameHcloudProviderOwned is the tag name we use to differentiate
	// cluster-api-provider-hcloud owned components from other tooling that
	// uses NameKubernetesClusterPrefix
	NameHcloudProviderOwned = "cluster." + NameHcloudProviderPrefix

	// MachineNameTag tags related MachineNameTag
	MachineNameTagKey = "machine." + NameHcloudProviderPrefix + "name"

	// MachineTempalteHashTag tags server resources
	MachineTemplateHashTagKey = "machine." + NameHcloudProviderPrefix + "template"
)

// ClusterTagKey generates the key for resources associated with a cluster.
func ClusterTagKey(name string) string {
	return fmt.Sprintf("%s%s", NameHcloudProviderOwned, name)
}

// ClusterHcloudCloudProviderTagKey generates the key for resources associated a cluster's Hcloud cloud provider.
func ClusterHcloudCloudProviderTagKey(name string) string {
	return fmt.Sprintf("%s%s", NameKubernetesHcloudCloudProviderPrefix, name)
}
