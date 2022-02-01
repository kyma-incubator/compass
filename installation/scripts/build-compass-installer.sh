#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=${CURRENT_DIR}/../../
IMAGE_NAME="$(${CURRENT_DIR}/extract-compass-installer-image.sh)"
BUILD_ARG=""

echo "
################################################################################
# Compass-Installer build
################################################################################
"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --installer-version)
            BUILD_ARG="--build-arg INSTALLER_VERSION=$2"
            shift
            shift
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

pushd $ROOT_DIR

docker build -t ${IMAGE_NAME} ${BUILD_ARG} -f ./tools/compass-installer/compass.Dockerfile .
docker push ${IMAGE_NAME}

popd 
