#!/usr/bin/env bash

set -eu -o pipefail

eval $(cat bazel-out/stable-status.txt | tr " " "=")

KUBECTL=$1
shift
KUSTOMIZE=$1
shift

WORK_DIR=`mktemp -d -p "$(pwd)"`
function cleanup {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

tar xf "${KUSTOMIZE}" -C "${WORK_DIR}"

cat > "${WORK_DIR}/image_bazel.json" <<EOF
[
  {"op": "replace", "path": "/spec/template/spec/containers/1/image", "value": "${STABLE_DOCKER_REGISTRY}/cluster-api-provider-hcloud:${STABLE_DOCKER_TAG}"}
]
EOF

cat >> "${WORK_DIR}/kustomization.yaml" <<EOF
bases:
- config

patchesJson6902:
- path: image_bazel.json
  target:
    group: apps
    version: v1
    kind: Deployment
    name: controller-manager
    namespace: capi-hcloud-system
- path: image_bazel.json
  target:
    group: apps
    version: v1
    kind: Deployment
    name: controller-manager
    namespace: capi-webhook-system
EOF

$KUBECTL kustomize "${WORK_DIR}/"
