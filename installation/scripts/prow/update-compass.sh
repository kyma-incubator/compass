#!/bin/bash

###
# Following script update compass installation only.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

docker exec k3d-kyma-agent-0  sh -c "ctr images rm \$(ctr image list -q name~=eu.gcr.io/kyma-project/incubator)"
docker exec k3d-kyma-server-0  sh -c "ctr images rm \$(ctr image list -q name~=eu.gcr.io/kyma-project/incubator)"

sudo ${INSTALLATION_DIR}/cmd/run.sh --skip-k3d-start --skip-kyma-start --skip-db-install
