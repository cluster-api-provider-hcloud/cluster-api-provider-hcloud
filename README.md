# cluster-api-provider-hcloud

Cluster API infrastructure provider for Hetzner Cloud https://hetzner.cloud

## Quick start

*More information available in the [Cluster API - Quick Start guide]*

- Make sure you have a Kubernetes management cluster available and your
  KUBECONFIG and context set correctly

- Ensure you have a recent [clusterctl] release (tested with v0.3.6)

- Ensure your Hcloud API token is created as secret in the kubernetes API:

```sh
kubectl create secret generic hcloud-token --from-literal=token=$TOKEN
```

- Register this infrastructure provider in your `$HOME/.cluster-api/clusterctl.yaml`:

```yaml
providers:
  - name: "hcloud"
    url: "https://github.com/simonswine/cluster-api-provider-hcloud/releases/latest/infrastructure-components.yaml"
    type: "InfrastructureProvider"
```

- Deploy the cluster API components to the management cluster

```sh
clusterctl init --infrastructure hcloud:v0.1.0-rc.3
```

- Create your first cluster `cluster-dev`

```sh
# The location of the cluster (fsn1|hel1|nbg1)
export HCLOUD_LOCATION=fsn1
# Name of SSH keys that have access to the cluster
export HCLOUD_SSH_KEY_NAME=id_rsa
# Instance types used (cf. https://www.hetzner.com/cloud)
export HCLOUD_NODE_MACHINE_TYPE=cx21
export HCLOUD_CONTROL_PLANE_MACHINE_TYPE=cx21

# Create cluster yamls
clusterctl config cluster cluster-dev --kubernetes-version v1.17.6 --control-plane-machine-count=1 --worker-machine-count=3 > cluster-dev.yaml

# Apply the resources
kubectl apply -f cluster-dev.yaml

# Watch resoruces being created
watch kubectl get hcloudclusters,cluster,hcloudmachines,machines,kubeadmcontrolplane
```

- Once the Control Plane has a ready replica, create a kubeconfig for the
  hcloud cluster and test connectivity:

```sh
KUBECONFIG_GUEST=$(pwd)/.kubeconfig-cluster-dev
kubectl get secrets cluster-dev-kubeconfig -o json | jq -r .data.value | base64 -d > $KUBECONFIG_GUEST
KUBECONFIG=$KUBECONFIG_GUEST kubectl get nodes,pods -A
```
[clusterctl]: https://github.com/kubernetes-sigs/cluster-api/releases/tag/v0.3.6
[Cluster API - Quick Start guide]: https://cluster-api.sigs.k8s.io/user/quick-start.html
