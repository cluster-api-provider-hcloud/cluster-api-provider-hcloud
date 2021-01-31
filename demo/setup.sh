#!/usr/bin/env bash

set -euo pipefail

set -x

kind create cluster --name capi-hcloud || true
export EXP_CLUSTER_RESOURCE_SET=true
clusterctl init --core cluster-api:v0.3.13