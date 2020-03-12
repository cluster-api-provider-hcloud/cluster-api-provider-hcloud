# cluster-api-provider-hetzner

## Quick start

- Deploy kind cluster + cluster api prerequisites using `./demo/setup.sh`
- Run controller `bazel run //cmd/cluster-api-provider-hetzner:run`
- Fill out token and image in `./demo/cluster-dev-*.yaml` deploy to kind
- Once cluster master is up get kubeadm's kubeconfig manually and apply ./manifests/config.jsonnet manually using kubecfg
