#!/usr/bin/env bash
# Copyright 2019 The Jetstack cert-manager contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

ROOT="$(dirname "${BASH_SOURCE[0]}")/.."
if [[ "$(basename $PWD)" == "__main__" ]]; then
    ROOT=$PWD
fi

source "${ROOT}/hack/lib/util.sh"

CMD=${1:-verify}

util::ensure_bazel $CMD "crds" "generated CRDs"

gazelle=$(realpath "$2")
kazel=$(realpath "$3")

go=$(realpath "$2")
controllergen="$(realpath "$3")"
export PATH=$(dirname "$go"):$PATH
export HOME=${HOME-$(realpath "$TEST_TMPDIR/home")}

util::before_job $CMD

set -o xtrace
"$controllergen" \
   object:headerFile=./hack/boilerplate.go.txt paths="./..."

"$controllergen" \
   crd:crdVersions=v1 rbac:roleName=manager-role paths="./..." output:crd:dir=config/crd/bases
set +o xtrace

util::after_job $CMD
