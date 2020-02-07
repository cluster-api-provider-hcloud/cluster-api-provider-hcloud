# cluster-api-provider-hetzner

## TODO

- Packer build automation
- Manifest deploy automation
- Report back APIServer URL so workers can join
- Attach floating IP to first control plane and add it's IPs to the certificate

## Usage

- Build image with packer
- Deploy kind cluster + prerequisites `./demo/setup.sh`
- Run controller `bazel run //cmd/cluster-api-provider-hetzner:run`
- Fill out token and image in `./demo/cluster-dev-*.yaml` deploy to kind
- Once cluster master is up get kubeadm's kubeconfig manually and apply ./manifests/config.jsonnet manually using kubecfg
