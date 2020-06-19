# Run a development build

## Prerequisites

This is just a really draft list what is required

* make
* bazel
* kind
* docker
* kubectl

## Running a development version

```bash
# Deploy kind cluster with cluster-api core componets
./demo/setup.sh

# Build project and deploy to local cluster
bazel run //cmd/cluster-api-provider-hcloud:deploy
```
