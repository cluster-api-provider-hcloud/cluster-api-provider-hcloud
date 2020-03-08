#!/bin/bash

set -euo pipefail

set -x

kind create cluster --name capi-hetzner || true

# Install cluster api manager
kubectl apply -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.2.10/cluster-api-components.yaml

# Install kubeadm bootstrap provider
kubectl apply -f https://github.com/kubernetes-sigs/cluster-api-bootstrap-provider-kubeadm/releases/download/v0.1.6/bootstrap-components.yaml
