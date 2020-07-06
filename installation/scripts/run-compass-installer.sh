#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"
COMPASS_OVERRIDES="${RESOURCES_DIR}/installer-overrides-compass.yaml"
COMPASS_CR="${RESOURCES_DIR}/installer-cr-compass-dependencies.yaml"
TMP_DIR="${CURRENT_DIR}/tmp-compass"

function cleanup {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

echo "
################################################################################
# Prepare Compass artifacts
################################################################################
"

mkdir -p "$TMP_DIR"
readonly RELEASE=$(<"${RESOURCES_DIR}"/COMPASS_VERSION)
curl -L "https://storage.googleapis.com/kyma-development-artifacts/compass/${RELEASE}/compass-installer.yaml" -o "${TMP_DIR}/compass-installer.yaml"
curl -L "https://storage.googleapis.com/kyma-development-artifacts/compass/${RELEASE}/is-installed.sh" -o "${TMP_DIR}/is-compass-installed.sh"

sed -i.bak '/action: install/d' "${TMP_DIR}/compass-installer.yaml"
COMBO_YAML=$(bash ${CURRENT_DIR}/concat-yamls.sh ${COMPASS_OVERRIDES} ${TMP_DIR}/compass-installer.yaml ${COMPASS_CR})
MINIKUBE_IP=$(minikube ip)
COMBO_YAML=$(sed 's/\.minikubeIP: .*/\.minikubeIP: '"${MINIKUBE_IP}"'/g' <<<"$COMBO_YAML")

echo "
################################################################################
# Install Compass version ${RELEASE}
################################################################################
"

kubectl create ns compass-installer
kubectl apply -f - <<< "$COMBO_YAML"
bash "${TMP_DIR}/is-compass-installed.sh"
