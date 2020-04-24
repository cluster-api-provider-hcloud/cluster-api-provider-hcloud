#!/usr/bin/env bash

set -eu -o pipefail

KUBECTL=$1
shift
KUSTOMIZE=$1
shift
IMAGE_TAR=$1
shift

WORK_DIR=`mktemp -d -p "$(pwd)"`
function cleanup {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

tar xf "${KUSTOMIZE}" -C "${WORK_DIR}"

cat > "${WORK_DIR}/image_bazel.json" <<EOF
[
  {"op": "replace", "path": "/spec/template/spec/containers/1/image", "value": "sha256:$(cat "${IMAGE_TAR/.tar/}.0.config.sha256")"}
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

kind load --name capi-hcloud -v 10 image-archive "${IMAGE_TAR}"

$KUBECTL apply -k "${WORK_DIR}/"
