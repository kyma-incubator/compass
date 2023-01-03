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

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
         --kyma-installation)
            checkInputParameterValue "${2}"
            KYMA_INSTALLATION="$2"
            shift
            shift
            ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *) # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

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

KYMA_COMPONENTS_FULL="${ROOT_PATH}"/installation/resources/kyma/kyma-components-full.yaml
KYMA_OVERRIDES_FULL="${ROOT_PATH}"/installation/resources/kyma/kyma-overrides-full.yaml

FULL_OVERRIDES_TEMP=overrides-full.yaml
cp ${KYMA_OVERRIDES_FULL} ${FULL_OVERRIDES_TEMP}

yq -i ".istio.helmValues.pilot.jwksResolverExtraRootCA = \"$CERT\"" "${FULL_OVERRIDES_TEMP}"
yq -i ".ory.oathkeeper.oathkeeper.config.authenticators.jwt.config.jwks_urls |= . + [\"$JWKS_URL\"]" "${FULL_OVERRIDES_TEMP}"

if [[ $(uname -m) == 'arm64' ]]; then
  yq -i ".istio.global.images.istio_proxyv2.containerRegistryPath = \"europe-west1-docker.pkg.dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_proxyv2.directory = \"sap-cp-cmp-dev/ucl-dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_proxyv2.version = \"1.13.2-distroless\"" "${FULL_OVERRIDES_TEMP}"

  yq -i ".istio.global.images.istio_pilot.containerRegistryPath = \"europe-west1-docker.pkg.dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_pilot.directory = \"sap-cp-cmp-dev/ucl-dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio_pilot.version = \"1.13.2-distroless\"" "${FULL_OVERRIDES_TEMP}"
fi

trap "rm -f ${MINIMAL_OVERRIDES_TEMP} ${FULL_OVERRIDES_TEMP}" EXIT INT TERM

KYMA_SOURCE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)

echo "Using Kyma source ${KYMA_SOURCE}"

# TODO: Remove after adoption of Kyma 2.4.3
KYMA_WORKSPACE=${HOME}/.kyma/sources/${KYMA_SOURCE}
if [[ -d "$KYMA_WORKSPACE" ]]
then
    echo "Kyma ${KYMA_SOURCE} already exists locally. Will attempt to sync it with remote..."
    rm -rf "$KYMA_WORKSPACE"/installation/resources/crds/service-catalog || true
    rm -rf "$KYMA_WORKSPACE"/installation/resources/crds/service-catalog-addons || true
else
    echo "Pulling Kyma ${KYMA_SOURCE}"
    git clone --single-branch --branch "${KYMA_SOURCE}" https://github.com/kyma-project/kyma.git "$KYMA_WORKSPACE"
fi

if [[ $KYMA_INSTALLATION == *full* ]]; then
  echo "Installing full Kyma"
  kyma deploy --components-file $KYMA_COMPONENTS_FULL --values-file $FULL_OVERRIDES_TEMP --source=local --workspace "$KYMA_WORKSPACE"
else
  echo "Installing minimal Kyma"
  kyma deploy --components-file $KYMA_COMPONENTS_MINIMAL  --values-file $MINIMAL_OVERRIDES_TEMP --source=local --workspace "$KYMA_WORKSPACE"
fi
