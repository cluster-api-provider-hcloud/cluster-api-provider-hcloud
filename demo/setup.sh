#!/usr/bin/env bash

set -euo pipefail

set -x

kind create cluster --name capi-hcloud || true

# Install cert-manager
kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.14.2/cert-manager.yaml

# Wait for cert-manager
kubectl wait -n cert-manager deployment cert-manager-webhook --for=condition=Available --timeout=120s

# Install cluster api components
kubectl apply \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.6/core-components.yaml \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.6/cluster-api-components.yaml \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.6/control-plane-components.yaml \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.6/bootstrap-components.yaml
