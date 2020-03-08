#!/bin/bash

set -euo pipefail

set -x

kind create cluster --name capi-hetzner || true

# Install cluster api manager
kubectl apply -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.2.10/cluster-api-components.yaml

# Install kubeadm bootstrap provider
kubectl apply -f https://github.com/kubernetes-sigs/cluster-api-bootstrap-provider-kubeadm/releases/download/v0.1.6/bootstrap-components.yaml

# Allow access to our CRDs
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: capi-hcloud
rules:
- apiGroups:
  - cluster-api-provider-hcloud.swine.dev
  resources:
  - hcloudclusters
  - hcloudmachines
  - hcloudtemplates
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: capi-hcloud
subjects:
- kind: ServiceAccount
  name: default
  namespace: capi-system
roleRef:
  kind: ClusterRole
  name: capi-hcloud
  apiGroup: rbac.authorization.k8s.io
EOF
