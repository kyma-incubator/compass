#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..
defaultRelease=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)
KYMA_RELEASE=${1:-$defaultRelease}
INSTALLER_CR_PATH="${ROOT_PATH}"/installation/resources/installer-cr-kyma-dependencies.yaml
OVERRIDES_COMPASS_GATEWAY="${ROOT_PATH}"/installation/resources/installer-overrides-compass-gateway.yaml
ISTIO_OVERRIDES="${ROOT_PATH}"/installation/resources/installer-overrides-istio.yaml
API_GATEWAY_OVERRIDES="${ROOT_PATH}"/installation/resources/installer-overrides-api-gateway.yaml

if [[ $KYMA_RELEASE == *PR-* ]] || [ $KYMA_RELEASE == master ]; then
  KYMA_TAG=$(curl -L https://storage.googleapis.com/kyma-development-artifacts/${KYMA_RELEASE}/kyma-installer-cluster.yaml | grep 'image: eu.gcr.io/kyma-project/kyma-installer:'| sed 's+image: eu.gcr.io/kyma-project/kyma-installer:++g' | tr -d '[:space:]')
else
  KYMA_TAG=${KYMA_RELEASE}
fi
kyma provision minikube
kyma install -c $INSTALLER_CR_PATH  -o $OVERRIDES_COMPASS_GATEWAY -o $ISTIO_OVERRIDES -o $API_GATEWAY_OVERRIDES --source "eu.gcr.io/kyma-project/kyma-installer:${KYMA_TAG}"

bash "${ROOT_PATH}"/installation/scripts/run-compass-installer.sh
bash "${ROOT_PATH}"/installation/scripts/is-installed.sh

echo "Adding Compass entries to /etc/hosts..."
sudo sh -c 'echo "\n$(minikube ip) adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local" >> /etc/hosts'
