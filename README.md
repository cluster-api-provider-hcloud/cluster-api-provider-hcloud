# cluster-api-provider-hcloud

Cluster API infrastructure provider for Hetzner Cloud https://hetzner.cloud

## Quick start

- Deploy kind cluster + cluster api prerequisites using `./demo/setup.sh`
- Deploy hcloud infra-provider: `kubectl apply -f https://github.com/simonswine/cluster-api-provider-hcloud/releases/download/v0.1.0-rc.1/manifests.yaml`
- Create the token secret in the kubernetes API: `kubectl create secret generic hcloud-token --from-literal=token=$my-token`
- Create cluster `kubectl apply -f https://raw.githubusercontent.com/simonswine/cluster-api-provider-hcloud/v0.1.0-rc.1/demo/cluster-dev.yaml`
- Create control plane `kubectl apply -f https://raw.githubusercontent.com/simonswine/cluster-api-provider-hcloud/v0.1.0-rc.1/demo/cluster-dev-controlplane.yaml`
- Create worker nodes `kubectl apply -f https://raw.githubusercontent.com/simonswine/cluster-api-provider-hcloud/v0.1.0-rc.1/demo/cluster-dev-worker.yaml`
- Watch meanwhile `watch kubectl get hcloudclusters,cluster,hcloudmachines,machines`
- Create kubeconfig for hcloud cluster and test connectivity:

```sh

kubectl get secrets cluster-dev-kubeconfig -o json | jq -r .data.value | base64 -d > .kubeconfig-hcloud
KUBECONFIG=.kubeconfig-hcloud kubectl get nodes,pods -A

```
