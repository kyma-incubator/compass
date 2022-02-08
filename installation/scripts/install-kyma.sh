#!/usr/bin/env bash

set -o errexit

echo "Installing Kyma..."

LOCAL_ENV=${LOCAL_ENV:-false}

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
OVERRIDES_DIR="${CURRENT_DIR}/../resources/kyma"
source $SCRIPTS_DIR/utils.sh

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --kyma-release)
            checkInputParameterValue "${2}"
            KYMA_RELEASE="$2"
            shift
            shift
            ;;
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

INSTALLER_CR_PATH="${ROOT_PATH}"/installation/resources/kyma/installer-cr-kyma-minimal.yaml
OVERRIDES_KYMA_MINIMAL_CFG_LOCAL="${ROOT_PATH}"/installation/resources/kyma/installer-overrides-kyma-minimal-config-local.yaml

MINIMAL_OVERRIDES_FILENAME=override-local-minimal.yaml
MINIMAL_OVERRIDES_CONTENT=$(sed "s~\"__CERT__\"~\"$CERT\"~" "${OVERRIDES_KYMA_MINIMAL_CFG_LOCAL}")

>"${MINIMAL_OVERRIDES_FILENAME}" cat <<-EOF
$MINIMAL_OVERRIDES_CONTENT
EOF

INSTALLER_CR_FULL_PATH="${ROOT_PATH}"/installation/resources/kyma/installer-cr-kyma.yaml
OVERRIDES_KYMA_FULL_CFG_LOCAL="${ROOT_PATH}"/installation/resources/kyma/installer-overrides-kyma-full-config-local.yaml

FULL_OVERRIDES_FILENAME=override-local-full.yaml
FULL_OVERRIDES_CONTENT=$(sed "s~\"__CERT__\"~\"$CERT\"~" "${OVERRIDES_KYMA_FULL_CFG_LOCAL}")

>"${FULL_OVERRIDES_FILENAME}" cat <<-EOF
$FULL_OVERRIDES_CONTENT
EOF

trap "rm -f ${MINIMAL_OVERRIDES_FILENAME} ${FULL_OVERRIDES_FILENAME}" EXIT INT TERM

if [[ $KYMA_RELEASE == *PR-* ]]; then
  KYMA_TAG=$(curl -L https://storage.googleapis.com/kyma-development-artifacts/${KYMA_RELEASE}/kyma-installer-cluster.yaml | grep 'image: eu.gcr.io/kyma-project/kyma-installer:'| sed 's+image: eu.gcr.io/kyma-project/kyma-installer:++g' | tr -d '[:space:]')
  if [ -z "$KYMA_TAG" ]; then echo "ERROR: Kyma artifacts for ${KYMA_RELEASE} not found."; exit 1; fi
  KYMA_SOURCE="eu.gcr.io/kyma-project/kyma-installer:${KYMA_TAG}"
elif [[ $KYMA_RELEASE == main ]]; then
  KYMA_SOURCE="main"
elif [[ $KYMA_RELEASE == *main-* ]]; then
  KYMA_SOURCE=$(echo $KYMA_RELEASE | sed 's+main-++g' | tr -d '[:space:]')
else
  KYMA_SOURCE="${KYMA_RELEASE}"
fi

echo "Using Kyma source ${KYMA_SOURCE}"

if [[ $KYMA_INSTALLATION == *full* ]]; then # todo add overrides if possible
  echo "Installing full Kyma"
  kyma deploy --components-file $INSTALLER_CR_FULL_PATH --source $KYMA_SOURCE
else
  echo "Installing minimal Kyma"
  kyma deploy --components-file $INSTALLER_CR_PATH  --values-file $OVERRIDES_DIR/kyma2.0-overrides.yaml --source $KYMA_SOURCE
fi