#!/usr/bin/env bash

set -o errexit

echo "Installing Kyma..."

LOCAL_ENV=${LOCAL_ENV:-false}

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "$CURRENT_DIR"/utils.sh

# $KYMA will take on the value "kyma" if not overridden - in local installation `run.sh` will override it
# This is done to make sure that `install-kyma.sh` can be used with different configurations and not only the local one
: ${KYMA:=kyma}
# $KUBECTL will take on the value "kubectl" if not overridden - in local installation `run.sh` will override it
# This is done to make sure that `install-kyma.sh` can be used with different configurations and not only the local one
: ${KUBECTL:=kubectl}

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

CERT=$(docker exec k3d-kyma-server-0 cat /var/lib/rancher/k3s/server/tls/server-ca.crt)
CERT="${CERT//$'\n'/\\\\n}"

KYMA_COMPONENTS_MINIMAL="${ROOT_PATH}"/installation/resources/kyma/kyma-components-minimal.yaml
KYMA_OVERRIDES_MINIMAL="${ROOT_PATH}"/installation/resources/kyma/kyma-overrides-minimal.yaml

MINIMAL_OVERRIDES_TEMP=overrides-minimal.yaml
cp ${KYMA_OVERRIDES_MINIMAL} ${MINIMAL_OVERRIDES_TEMP}

yq -i ".istio.helmValues.pilot.jwksResolverExtraRootCA = \"$CERT\"" "${MINIMAL_OVERRIDES_TEMP}"

trap "rm -f ${MINIMAL_OVERRIDES_TEMP}" EXIT INT TERM

KYMA_SOURCE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)

echo "Using Kyma source ${KYMA_SOURCE}"

# Reuse Kyma source, otherwise the Kyma source is fetched everytime
KYMA_WORKSPACE=${HOME}/.kyma/sources/${KYMA_SOURCE}
if [[ -d "$KYMA_WORKSPACE" ]]
then
   echo "Kyma ${KYMA_SOURCE} already exists locally."
else
   echo "Pulling Kyma ${KYMA_SOURCE}"
   git clone --single-branch --branch "${KYMA_SOURCE}" https://github.com/kyma-project/kyma.git "$KYMA_WORKSPACE"
fi

echo "Installing minimal Kyma"
"$KYMA" deploy --components-file "$KYMA_COMPONENTS_MINIMAL"  --values-file "$MINIMAL_OVERRIDES_TEMP" --source=local --workspace "$KYMA_WORKSPACE"
