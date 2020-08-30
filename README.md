# cluster-api-provider-hcloud

Cluster API infrastructure provider for Hetzner Cloud https://hetzner.cloud

## Docs

https://docs.capihc.com/
or under ./docs/src

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
    url: "https://github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/releases/latest/infrastructure-components.yaml"
    type: "InfrastructureProvider"
```

- Deploy the cluster API components to the management cluster

```sh
clusterctl init --infrastructure hcloud:v0.1.0-rc.4
```

- Create your first cluster `cluster-dev`

```sh
# The location of the cluster (fsn1|hel1|nbg1)
export HCLOUD_LOCATION=fsn1
# Name of SSH keys that have access to the cluster, you need to upload them before
export HCLOUD_SSH_KEY_NAME=id_rsa
# Instance types used (cf. https://www.hetzner.com/cloud) 
# Caution! Do not use a cx11 for control-plane! Kubadm requires more than 1 vCPU
export HCLOUD_NODE_MACHINE_TYPE=cx21
export HCLOUD_CONTROL_PLANE_MACHINE_TYPE=cx21

# Create cluster yamls
clusterctl config cluster cluster-dev --kubernetes-version v1.18.3 --control-plane-machine-count=1 --worker-machine-count=3 > cluster-dev.yaml

# Apply the resources
kubectl apply -f cluster-dev.yaml

# Watch resources being created
watch -n 1 kubectl get hcloudclusters,cluster,hcloudmachines,machines,kubeadmcontrolplane
```
The cluster need some time until it is ready:
| Task | Time |
| ---- | ---- |
| Full cluster | ~20-25min
| For the packer | ~10-15min
| Snapshot | ~2-3min 
| Cluster creation without packer and snapshot 3 control planes, 3 worker | ~10min
| Worker upscale | ~1-2min
| Worker downscale |  ~20s
| Control plane upscale per node | ~2.5min
| Control plane downscale per node | ~1min


- Once the Control Plane has a ready replica, create a kubeconfig for the
  hcloud cluster and test connectivity:

```sh
KUBECONFIG_GUEST=$(pwd)/.kubeconfig-cluster-dev
kubectl get secrets cluster-dev-kubeconfig -o json | jq -r .data.value | base64 -d > $KUBECONFIG_GUEST
KUBECONFIG=$KUBECONFIG_GUEST kubectl get nodes,pods -A
```
[clusterctl]: https://github.com/kubernetes-sigs/cluster-api/releases/tag/v0.3.6
[Cluster API - Quick Start guide]: https://cluster-api.sigs.k8s.io/user/quick-start.html


## For Developers or demo purpose
See ./docs/src/developers or https://docs.capihc.com/developer/developer.html

### Prerequisites

- clusterctl
- docker
- kind
- kubectl
- kustomize
- kubebuilder
- packer
- BAZEL
- Go 1.13
- gomock
- watch (On MAC: `brew install watch`)
- JQ (On MAC: `brew install jq`)

- Running development version

```sh
# Deploy kind cluster with cluster-api core componets
./demo/setup.sh

# Build project and deploy to local cluster
bazel run //cmd/cluster-api-provider-hcloud:deploy
```

- Applying the target cluster with demo-cluster

```sh
# Create the 3 SSH Keys and name the keys control-plane, worker and cluster
ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/<control-plane | worker | cluster>

# Create a Project on Hetzner Cloud and upload them

# Create a token on Hetzner Cloud and apply it as secret
kubectl create secret generic hcloud-token --from-literal=token=$TOKEN

# Apply the manifest to your management cluster; cluster name is cluster-dev; use quickstart guide for getting access to the target cluster
kubectl apply -f ./demo/demo-cluster.yaml

# Deleting the cluster
kubectl delete -f ./demo/demo-cluster.yaml

# Deleting the management cluster
kind delete cluster --name capi-hcloud
```

