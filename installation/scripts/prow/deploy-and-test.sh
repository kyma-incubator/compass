#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma with Compass on Minikube, and runs the integrations tests.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

export ARTIFACTS="/var/log/prow_artifacts"
sudo mkdir -p "${ARTIFACTS}"

sudo ${INSTALLATION_DIR}/cmd/run.sh
sudo ARTIFACTS=${ARTIFACTS} ${INSTALLATION_DIR}/scripts/testing.sh
