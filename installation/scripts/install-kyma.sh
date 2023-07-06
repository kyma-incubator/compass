#!/usr/bin/env bash

set -o errexit

echo "Installing Kyma..."

LOCAL_ENV=${LOCAL_ENV:-false}

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
OVERRIDES_DIR="${CURRENT_DIR}/../resources/kyma"
DEFAULT_JWKS_URL="http://ory-hydra-public.kyma-system.svc.cluster.local:4444/.well-known/jwks.json"
source $SCRIPTS_DIR/utils.sh

usek3d

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

CERT=$(docker exec k3d-kyma-server-0 cat /var/lib/rancher/k3s/server/tls/server-ca.crt)
CERT="${CERT//$'\n'/\\\\n}"

VALUES_FILE="${ROOT_PATH}"/chart/compass/values.yaml
IDP_HOST=$(yq ".global.cockpit.auth.idpHost" $VALUES_FILE)
AUTH_PATH=$(yq ".global.cockpit.auth.path" $VALUES_FILE)

if [ -z "$IDP_HOST" ]; then
  JWKS_URL=$DEFAULT_JWKS_URL # in case no oidc args were passed nor config file
else
  JWKS_URL=$IDP_HOST$AUTH_PATH
fi

KYMA_COMPONENTS_MINIMAL="${ROOT_PATH}"/installation/resources/kyma/kyma-components-minimal.yaml
KYMA_OVERRIDES_MINIMAL="${ROOT_PATH}"/installation/resources/kyma/kyma-overrides-minimal.yaml

MINIMAL_OVERRIDES_TEMP=overrides-minimal.yaml
cp ${KYMA_OVERRIDES_MINIMAL} ${MINIMAL_OVERRIDES_TEMP}

yq -i ".istio.helmValues.pilot.jwksResolverExtraRootCA = \"$CERT\"" "${MINIMAL_OVERRIDES_TEMP}"
yq -i ".ory.oathkeeper.oathkeeper.config.authenticators.jwt.config.jwks_urls |= . + [\"$JWKS_URL\"]" "${MINIMAL_OVERRIDES_TEMP}"

if [[ $(uname -m) == 'arm64' ]]; then
  yq -i ".istio.global.images.istio_proxyv2.containerRegistryPath = \"europe-west1-docker.pkg.dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_proxyv2.directory = \"sap-cp-cmp-dev/ucl-dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_proxyv2.version = \"1.13.2-distroless\"" "${MINIMAL_OVERRIDES_TEMP}"

  yq -i ".istio.global.images.istio_pilot.containerRegistryPath = \"europe-west1-docker.pkg.dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_pilot.directory = \"sap-cp-cmp-dev/ucl-dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_pilot.version = \"1.13.2-distroless\"" "${MINIMAL_OVERRIDES_TEMP}"
fi

trap "rm -f ${MINIMAL_OVERRIDES_TEMP}" EXIT INT TERM

KYMA_SOURCE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)

echo "Using Kyma source ${KYMA_SOURCE}"

# TODO: Remove after adoption of Kyma 2.4.3 and change kyma deploy command source to --source $KYMA_SOURCE
KYMA_WORKSPACE=${HOME}/.kyma/sources/${KYMA_SOURCE}
if [[ -d "$KYMA_WORKSPACE" ]]
then
   echo "Kyma ${KYMA_SOURCE} already exists locally."
else
   echo "Pulling Kyma ${KYMA_SOURCE}"
   git clone --single-branch --branch "${KYMA_SOURCE}" https://github.com/kyma-project/kyma.git "$KYMA_WORKSPACE"
fi

rm -rf "$KYMA_WORKSPACE"/installation/resources/crds/service-catalog || true
rm -f "$KYMA_WORKSPACE"/installation/resources/crds/service-catalog-addons/clusteraddonsconfigurations.addons.crd.yaml || true
rm -f "$KYMA_WORKSPACE"/installation/resources/crds/service-catalog-addons/addonsconfigurations.addons.crd.yaml || true

echo "Installing minimal Kyma"
kyma deploy --components-file $KYMA_COMPONENTS_MINIMAL  --values-file $MINIMAL_OVERRIDES_TEMP --source=local --workspace "$KYMA_WORKSPACE"

echo "Helm install ORY components"
helm upgrade --install ory -f "${ROOT_PATH}"/chart/ory/values.yaml -n kyma-system "${ROOT_PATH}"/chart/ory