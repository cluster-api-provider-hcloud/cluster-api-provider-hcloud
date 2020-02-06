#!/bin/sh

set -eu -o pipefail

CAPH=$1
shift
TAR=$1
shift

tar xf $TAR

export PATH=$(pwd)/usr/local/bin:$PATH

exec $CAPH "$@"
