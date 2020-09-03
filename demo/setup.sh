#!/usr/bin/env bash

set -euo pipefail

set -x

kind create cluster --name capi-hcloud || true

clusterctl init --core cluster-api:v0.3.9