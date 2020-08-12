#!/usr/bin/env bash

set -o errexit

echo "Installing Kyma..."

LOCAL_ENV=${LOCAL_ENV:-false}
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..
defaultRelease=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)
KYMA_RELEASE=${1:-$defaultRelease}
INSTALLER_CR_PATH="${ROOT_PATH}"/installation/resources/installer-cr-kyma-dependencies.yaml
OVERRIDES_COMPASS_GATEWAY="${ROOT_PATH}"/installation/resources/installer-overrides-compass-gateway.yaml
ISTIO_OVERRIDES="${ROOT_PATH}"/installation/resources/installer-overrides-istio.yaml
API_GATEWAY_OVERRIDES="${ROOT_PATH}"/installation/resources/installer-overrides-api-gateway.yaml

if [[ $KYMA_RELEASE == *PR-* ]]; then
  KYMA_TAG=$(curl -L https://storage.googleapis.com/kyma-development-artifacts/${KYMA_RELEASE}/kyma-installer-cluster.yaml | grep 'image: eu.gcr.io/kyma-project/kyma-installer:'| sed 's+image: eu.gcr.io/kyma-project/kyma-installer:++g' | tr -d '[:space:]')
  if [ -z "$KYMA_TAG" ]; then echo "ERROR: Kyma artifacts for ${KYMA_RELEASE} not found."; exit 1; fi
  KYMA_SOURCE="eu.gcr.io/kyma-project/kyma-installer:${KYMA_TAG}"
elif [[ $KYMA_RELEASE == master ]]; then
  KYMA_SOURCE="latest-published"
elif [[ $KYMA_RELEASE == *master-* ]]; then
  KYMA_SOURCE=$(echo $KYMA_RELEASE | sed 's+master-++g' | tr -d '[:space:]')
else
  KYMA_SOURCE="${KYMA_RELEASE}"
fi

echo "Using Kyma source '${KYMA_SOURCE}'..."

if [[ "$LOCAL_ENV" == "true" ]]; then
    echo "Setting overrides for local environment..."
    ADDITIONAL_PARAMS="-o ${ISTIO_OVERRIDES}${ADDITIONAL_PARAMS}"
fi

echo "Configuring Tiller..."
kubectl apply -f "${ROOT_PATH}"/installation/resources/tiller.yaml

echo "Installing Kyma..."
set -o xtrace
kyma install -c $INSTALLER_CR_PATH -o $OVERRIDES_COMPASS_GATEWAY -o $API_GATEWAY_OVERRIDES ${ADDITIONAL_PARAMS} --source $KYMA_SOURCE
set +o xtrace