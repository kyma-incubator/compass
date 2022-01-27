#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $CURRENT_DIR/utils.sh
#useMinikube

DOMAIN="kyma.local"

VM_DRIVER="virtualbox"
if [ `uname -s` = "Darwin" ]; then
    VM_DRIVER="hyperkit"
fi

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --skip-minikube-start)
            SKIP_MINIKUBE_START=true
            shift # past argument
        ;;
        --cr)
            checkInputParameterValue "$2"
            CR_PATH="$2"
            shift # past argument
            shift # past value
        ;;
        --vm-driver)
            checkInputParameterValue "$2"
            VM_DRIVER="$2"
            shift
            shift
        ;;
        --password)
            checkInputParameterValue "$2"
            ADMIN_PASSWORD="${2}"
            shift # past argument
            shift # past value
        ;;
        --kyma-installation)
            checkInputParameterValue "${2}"
            KYMA_INSTALLATION="${2}"
            shift # past argument
            shift # past value
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

bash ${SCRIPTS_DIR}/build-compass-installer.sh --vm-driver "${VM_DRIVER}"

if [ -z "$CR_PATH" ]; then

    TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
    CR_PATH="${TMPDIR}/installer-cr-local.yaml"
    bash ${SCRIPTS_DIR}/create-cr.sh --output "${CR_PATH}"

fi

bash ${SCRIPTS_DIR}/installer.sh --cr "${CR_PATH}" --password "${ADMIN_PASSWORD}" --kyma-installation "${KYMA_INSTALLATION}"
rm -rf $TMPDIR
