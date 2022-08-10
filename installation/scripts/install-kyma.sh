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
  yq -i ".istio.global.containerRegistry.path = \"europe-west1-docker.pkg.dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio.directory = \"sap-cp-cmp-dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio.name = \"ucl-dev\"" "${MINIMAL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio.version = \"1.11.4-distroless\"" "${MINIMAL_OVERRIDES_TEMP}"
fi

KYMA_COMPONENTS_FULL="${ROOT_PATH}"/installation/resources/kyma/kyma-components-full.yaml
KYMA_OVERRIDES_FULL="${ROOT_PATH}"/installation/resources/kyma/kyma-overrides-full.yaml

FULL_OVERRIDES_TEMP=overrides-full.yaml
cp ${KYMA_OVERRIDES_FULL} ${FULL_OVERRIDES_TEMP}

yq -i ".istio.helmValues.pilot.jwksResolverExtraRootCA = \"$CERT\"" "${FULL_OVERRIDES_TEMP}"
yq -i ".ory.oathkeeper.oathkeeper.config.authenticators.jwt.config.jwks_urls |= . + [\"$JWKS_URL\"]" "${FULL_OVERRIDES_TEMP}"

if [[ $(uname -m) == 'arm64' ]]; then
  yq -i ".istio.global.containerRegistry.path = \"europe-west1-docker.pkg.dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio.directory = \"sap-cp-cmp-dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio.name = \"ucl-dev\"" "${FULL_OVERRIDES_TEMP}"
  yq -i ".istio.global.images.istio.version = \"1.11.4-distroless\"" "${FULL_OVERRIDES_TEMP}"
fi

trap "rm -f ${MINIMAL_OVERRIDES_TEMP} ${FULL_OVERRIDES_TEMP}" EXIT INT TERM

KYMA_SOURCE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)

echo "Using Kyma source ${KYMA_SOURCE}"

if [[ $KYMA_INSTALLATION == *full* ]]; then
  echo "Installing full Kyma"
  kyma deploy --components-file $KYMA_COMPONENTS_FULL --values-file $FULL_OVERRIDES_TEMP --source $KYMA_SOURCE
else
  echo "Installing minimal Kyma"
  kyma deploy --components-file $KYMA_COMPONENTS_MINIMAL  --values-file $MINIMAL_OVERRIDES_TEMP --source $KYMA_SOURCE
fi
