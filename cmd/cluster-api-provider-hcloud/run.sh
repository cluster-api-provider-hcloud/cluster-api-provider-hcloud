#!/usr/bin/env bash

set -eu -o pipefail

CAPH=$1
shift
TAR=$1
shift

tar xf $TAR

export PATH=$(pwd)/usr/local/bin:$PATH

exec $CAPH \
  --verbose \
  --manifests-config-path "./manifests-config/config-extvar.jsonnet" \
  --packer-config-path "./packer-config/packer-centos7-crio.json" \
  "$@"
