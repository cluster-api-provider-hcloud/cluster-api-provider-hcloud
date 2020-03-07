#!/bin/bash

set -euo pipefail

set -x

kind create cluster --name capi-hetzner || true

# Install cert-manager
kubectl create namespace cert-manager --dry-run -o yaml | kubectl apply -f -
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.14.2/cert-manager.yaml

# Wait for cert-manager
kubectl wait -n cert-manager deployment cert-manager-webhook --for=condition=Available --timeout=120s

# Install cluster api components
kubectl apply \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.3/core-components.yaml \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.3/cluster-api-components.yaml \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.3/control-plane-components.yaml \
  -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.3.3/bootstrap-components.yaml

# Allow access to our CRDs
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: capi-hcloud-clusters
  labels:
    cluster.x-k8s.io/provider: infrastructure-hcloud
    cluster.x-k8s.io/aggregate-to-manager: "true"
rules:
- apiGroups:
  - cluster-api-provider-hcloud.swine.dev
  resources:
  - hcloudclusters
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
kind: ClusterRole
metadata:
  name: capi-hcloud-machines
  labels:
    cluster.x-k8s.io/provider: infrastructure-hcloud
    cluster.x-k8s.io/aggregate-to-manager: "true"
rules:
- apiGroups:
  - cluster-api-provider-hcloud.swine.dev
  resources:
  - hcloudmachines
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
kind: ClusterRole
metadata:
  name: capi-hcloud-machine-templates
  labels:
    cluster.x-k8s.io/provider: infrastructure-hcloud
    cluster.x-k8s.io/aggregate-to-manager: "true"
rules:
- apiGroups:
  - cluster-api-provider-hcloud.swine.dev
  resources:
  - hcloudmachinetemplates
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
kind: ClusterRole
metadata:
  name: capi-kubeadm-control-plane-manager-hcloud
  labels:
    cluster.x-k8s.io/provider: infrastructure-hcloud
rules:
- apiGroups:
  - cluster-api-provider-hcloud.swine.dev
  resources:
  - hcloudmachinetemplates
  - hcloudmachines
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
  labels:
    cluster.x-k8s.io/provider: infrastructure-hcloud
  name: capi-kubeadm-control-plane-manager-hcloud
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: capi-kubeadm-control-plane-manager-hcloud
subjects:
- kind: ServiceAccount
  name: default
  namespace: capi-kubeadm-control-plane-system
EOF
