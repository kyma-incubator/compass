#!/bin/sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

defaultRelease="master"
KYMA_RELEASE=${1:-$defaultRelease}

kyma provision minikube
kyma install -o "${ROOT_PATH}"/installation/resources/installer-cr.yaml -o "${ROOT_PATH}"/installation/resources/installer-config.yaml --release "${KYMA_RELEASE}"