#!/bin/bash

###
# Following script update compass installation only.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

sudo ${INSTALLATION_DIR}/cmd/run.sh --skip-k3d-start --skip-kyma-start --skip-db-install
