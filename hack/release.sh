#!/usr/bin/env bash

set -eu -o pipefail

GHR=$1
shift
RELEASE_METADATA=$1
shift
RELEASE_TAR=$1
shift

WORK_DIR=`mktemp -d -p "$(pwd)"`
function cleanup {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

tar xf "${RELEASE_TAR}" -C "${WORK_DIR}"

eval $(cat ${RELEASE_METADATA})

BODY=$(cat <<-END

\`\`\`
$DOCKER_REGISTRY:${DOCKER_TAG}
$DOCKER_REGISTRY@${IMAGE_MANAGER_DIGEST}
\`\`\`

END
)

set -x

$GHR \
  -username simonswine \
  -repository cluster-api-provider-hcloud \
  -commitish "${GIT_COMMIT}" \
  -body "${BODY}" \
  -prerelease \
  -replace \
  "${SCM_REVISION}" \
  "${WORK_DIR}"
