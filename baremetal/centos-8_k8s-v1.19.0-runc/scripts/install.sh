#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail


source ./configure_base.sh
source ./configure_cri.sh
source ./configure_kubernetes.sh
source ./configure_kernel.sh