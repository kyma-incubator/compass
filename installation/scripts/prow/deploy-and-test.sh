#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma with Compass on Minikube, and runs the integrations tests.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

export ARTIFACTS="/var/log/prow_artifacts"
sudo mkdir -p "${ARTIFACTS}"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --dump-db)
            DUMP_DB=true
            shift # past argument
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters


if [[ ${DUMP_DB} ]]; then
    sudo ${INSTALLATION_DIR}/cmd/run.sh --k3d-memory 12288 --dump-db
else
    sudo ${INSTALLATION_DIR}/cmd/run.sh --k3d-memory 12288
fi

if [[ ${DUMP_DB} ]]; then
    sudo ARTIFACTS=${ARTIFACTS} ${INSTALLATION_DIR}/scripts/testing.sh --dump-db true
else
    sudo ARTIFACTS=${ARTIFACTS} ${INSTALLATION_DIR}/scripts/testing.sh --dump-db false
fi


