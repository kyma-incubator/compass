#!/bin/bash

###
# Following script installs necessary tooling for Debian, starts Director and Ord service, and runs the smoke tests.
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


#sudo ${INSTALLATION_DIR}/cmd/run.sh
echo "pwd:"
pwd
echo "CURRENT_DIR=${CURRENT_DIR}"
echo "INSTALLATION_DIR=${INSTALLATION_DIR}"
echo "ARTIFACTS=${ARTIFACTS}"
echo "ord-test end reached!"
