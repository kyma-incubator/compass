#!/usr/bin/env bash

set -o errexit

echo "Installing Kyma..."

LOCAL_ENV=${LOCAL_ENV:-false}

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
OVERRIDES_DIR="${CURRENT_DIR}/../resources/kyma"
source $SCRIPTS_DIR/utils.sh

usek3d

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

CERT=$(docker exec k3d-kyma-server-0 cat /var/lib/rancher/k3s/server/tls/server-ca.crt)
CERT="${CERT//$'\n'/\\\\n}"

KYMA_COMPONENTS_MINIMAL="${ROOT_PATH}"/installation/resources/kyma/kyma-components-minimal.yaml
KYMA_OVERRIDES_MINIMAL="${ROOT_PATH}"/installation/resources/kyma/kyma-overrides-minimal.yaml

MINIMAL_OVERRIDES_TEMP=overrides-minimal.yaml
cp ${KYMA_OVERRIDES_MINIMAL} ${MINIMAL_OVERRIDES_TEMP}

yq -i ".istio.helmValues.pilot.jwksResolverExtraRootCA = \"$CERT\"" "${MINIMAL_OVERRIDES_TEMP}"

if [[ $(uname -m) == 'arm64' ]]; then
  yq -i ".istio.global.images.istio_proxyv2.containerRegistryPath = \"ghcr.io\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_proxyv2.directory = \"resf/istio\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_proxyv2.version = \"1.14.1-distroless\"" "${MINIMAL_OVERRIDES_TEMP}"

  yq -i ".istio.global.images.istio_pilot.containerRegistryPath = \"ghcr.io\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_pilot.directory = \"resf/istio\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_pilot.version = \"1.14.1-distroless\"" "${MINIMAL_OVERRIDES_TEMP}"
fi

trap "rm -f ${MINIMAL_OVERRIDES_TEMP}" EXIT INT TERM

KYMA_SOURCE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)

echo "Using Kyma source ${KYMA_SOURCE}"

echo "Installing minimal Kyma"
kyma deploy --components-file $KYMA_COMPONENTS_MINIMAL  --values-file $MINIMAL_OVERRIDES_TEMP --source="$KYMA_SOURCE"
