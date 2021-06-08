#!/usr/bin/env bash

###
# Following script generates compass-installer artifacts for a release.
#
# INPUTS:
# - COMPASS_INSTALLER_PUSH_DIR - (optional) directory where kyma-installer docker image is pushed, if specified should ends with a slash (/)
# - COMPASS_INSTALLER_VERSION - version (image tag) of kyma-installer
# - ARTIFACTS_DIR - path to directory where artifacts will be stored
#
###

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
INSTALLER_YAML_PATH="${RESOURCES_DIR}/installer.yaml"
INSTALLER_LOCAL_CONFIG_PATH="${RESOURCES_DIR}/installer-config-local.yaml.tpl"
INSTALLER_CR_PATH="${RESOURCES_DIR}/installer-cr.yaml.tpl"

function generateArtifact() {
    TMP_CR=$(mktemp)

    ${CURRENT_DIR}/create-cr.sh --url "" --output "${TMP_CR}" --version 0.0.1 --crtpl_path "${INSTALLER_CR_PATH}"

    ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER_YAML_PATH} ${TMP_CR} \
      | sed -E ";s;image: eu.gcr.io\/kyma-project\/develop\/installer:.+;image: eu.gcr.io/kyma-project/${COMPASS_INSTALLER_PUSH_DIR}compass-installer:${COMPASS_INSTALLER_VERSION};" \
      > ${ARTIFACTS_DIR}/compass-installer.yaml

    cp ${INSTALLER_LOCAL_CONFIG_PATH} ${ARTIFACTS_DIR}/compass-config-local.yaml

    rm -rf ${TMP_CR}
}

function copyKymaInstaller() {
    release=$(<"${RESOURCES_DIR}"/KYMA_VERSION)
if [[ $release == *PR-* ]] || [[ $release == *master* ]]; then
    curl -L https://storage.googleapis.com/kyma-development-artifacts/${release}/kyma-installer-cluster.yaml -o kyma-installer.yaml
    curl -L https://storage.googleapis.com/kyma-development-artifacts/${release}/is-installed.sh -o ${ARTIFACTS_DIR}/is-kyma-installed.sh
else
    curl -L https://storage.googleapis.com/kyma-prow-artifacts/${release}/kyma-installer-cluster.yaml -o kyma-installer.yaml
    cp ${SCRIPTS_DIR}/is-kyma-installed.sh ${ARTIFACTS_DIR}/is-kyma-installed.sh
fi

    sed -i '/action: install/d' kyma-installer.yaml
    cat ${RESOURCES_DIR}/kyma/installer-cr-kyma-minimal.yaml >> kyma-installer.yaml
    mv kyma-installer.yaml ${ARTIFACTS_DIR}/kyma-installer.yaml
}

generateArtifact
copyKymaInstaller
