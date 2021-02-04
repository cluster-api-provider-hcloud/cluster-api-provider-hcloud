# cluster-api-provider-hcloud

Cluster API infrastructure provider for Hetzner Cloud https://hetzner.cloud

## Docs

https://docs.capihc.com/
or under ./docs/src

## Time estimation

| Task | Time |
| ---- | ---- |
| Full cluster | ~15-30min
| For the packer | ~10-15min
| Snapshot | ~2-3min 
| First control-plane, worker-nodes are created | after ~4min
| Cluster creation without packer and snapshot 3 control planes, 3 worker | ~10min
| Worker upscale | ~1-2min
| Worker downscale |  ~20s
| Control plane upscale per node | ~2.5min
| Control plane downscale per node | ~1min


## Quick start

*More information available in the [Cluster API - Quick Start guide]*

Before you can start you need a management Cluster.
If you have no management cluster you can use the ./demo/setup.sh to get a kind cluster. 
If you are not using the script because you have already a managment cluster please ensure to have the following enabled:

```sh
export EXP_CLUSTER_RESOURCE_SET=true
clusterctl init --core cluster-api:v0.3.13
```

Please ensure you have a recent [clusterctl] release (tested with v0.3.13). You can test with `clusterctl version`

Now we can start by creating a secret in management cluster. $TOKEN is a placeholder for your HETZNER API Token. You can create one in your Project under security/API TOKENS.

```sh
kubectl create secret generic hetzner-token --from-literal=token=$TOKEN
```

Then we need to create an SSH Key for the nodes. Because this is a quickstart we have specified the name of the Key, but of course feel free to change the name, but remember to do it also in cluster.yaml file.

```sh
ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/cluster
```

For deploying necessary applications like the CNI, CCM, CSI etc. We use the ClusterResourceSets and apply them to our managment cluster. 

```sh
kubectl apply -f ./demo/ClusterResourceSets
```

Then we need to register this infrastructure provider in your `$HOME/.cluster-api/clusterctl.yaml`:

```yaml
providers:
  - name: "hcloud"
    url: "https://github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/releases/latest/infrastructure-components.yaml"
    type: "InfrastructureProvider"
```

Now we deploy the API components to the management cluster

```sh
clusterctl init --infrastructure hcloud:v0.1.0
```

Now we can deploy our first Cluster. For production use it is recommended to use your own templates with all configurations. [name] is the placeholder for your cluster name like cluster-dev

```sh
clusterctl config cluster [name] | kubectl apply -f -

or use helm

helm install cluster ./demo/helm-charts/cluster-demo

```

You can check now the status of your target cluster via your management cluster:

```sh
kubectl get cluster --all-namespaces

### To verify the first control plane is up:
kubectl get kubeadmcontrolplane --all-namespaces
```
To get access to your target cluster you can retrieve the kubeconfig file and use it via ENV. [name] is the placeholder for your above defined cluster name.
```sh
export KUBECONFIG_GUEST=$(pwd)/.kubeconfig-[name]
kubectl --namespace=default get secret [name]-kubeconfig \
   -o jsonpath={.data.value} | base64 --decode \
   > $KUBECONFIG_GUEST
```

To verify you have access try:
```sh
KUBECONFIG=$KUBECONFIG_GUEST kubectl get nodes
```

If you want you can now move all the cluster-api Resources from your management Cluster to your Target Cluster:

```sh
export EXP_CLUSTER_RESOURCE_SET=true
KUBECONFIG=$KUBECONFIG_GUEST clusterctl init --core cluster-api:v0.3.13
KUBECONFIG=$KUBECONFIG_GUEST clusterctl init --infrastructure hcloud:v0.1.0
clusterctl move --to-kubeconfig $KUBECONFIG_GUEST
```

To delete the cluster (if management cluster not equal target cluster)

```sh
kubectl delete cluster [name]

or with helm

helm uninstall cluster
```

To delete your managment cluster (setup via setup.sh)
```sh
kind delete cluster --name capi-hcloud
```

# Debugging
```sh
### Getting information about the cluster
KUBECONFIG=$KUBECONFIG_GUEST kubectl get all,nodes -A

### Getting informations about cluster-api
watch kubectl get hcloudclusters,cluster,hcloudmachines,baremetalmachines,machines

### cluster-info
KUBECONFIG=$KUBECONFIG_GUEST kubectl get cm cluster-info -n kube-public -o yaml

# Logs
### Provider Integration
kubectl logs -f deployment/capi-hcloud-controller-manager -c manager -n capi-hcloud-system

### Cluster-API Controller
kubectl logs -f deployment/capi-controller-manager -c manager -n capi-system

### Bootstrap Controller
kubectl logs -f deployment/capi-kubeadm-bootstrap-controller-manager -c manager  -n capi-kubeadm-bootstrap-system

### Kubeadm Control-plane Controller
kubectl logs -f deployment/capi-kubeadm-control-plane-controller-manager -c manager  -n capi-kubeadm-control-plane-system

### Kubernetes Events
kubectl get events -o custom-columns=FirstSeen:.firstTimestamp,LastSeen:.lastTimestamp,Count:.count,From:.source.component,Type:.type,Re│
ason:.reason,Message:.message --watch

### Get kubeadm-config
kubectl -n kube-system get cm kubeadm-config -o yaml

```


[clusterctl]: https://github.com/kubernetes-sigs/cluster-api/releases/
[Cluster API - Quick Start guide]: https://cluster-api.sigs.k8s.io/user/quick-start.html


## For Developers 
> Please use this for testing!

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

This creates the management cluster with all the controllers
```sh
# Deploy kind cluster with cluster-api core componets
./demo/setup.sh

# Build project and deploy to local cluster
make deploy_kind
```

- Applying the target cluster with demo-cluster

```sh
# Please create an SSH Key for later access on the nodes.
ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/cluster

# Create a Project on Hetzner Cloud and upload the public key. 

# Create a token on Hetzner Cloud and apply it as secret
kubectl create secret generic hetzner-token --from-literal=token=$TOKEN

#For automatic installation of manifests we use ClusterResourceSets
kubectl apply -f demo/ClusterResourceSets

## You can choose which manifests should be applyed by setting the value of the labels under kind: Cluster

# Apply the manifest to your management cluster; use quickstart guide for getting access to the target cluster
kubectl apply -f ./demo/cluster-minimal.yaml

## Get Logs:
kubectl logs -f deployment/capi-hcloud-controller-manager -c manager --v=4 -n capi-hcloud-system

# Deleting the target cluster
kubectl delete -f ./demo/demo-cluster.yaml

# Deleting the controller
make delete_capihc

# Deleting the management cluster
kind delete cluster --name capi-hcloud

```


