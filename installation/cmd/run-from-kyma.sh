#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

defaultRelease="master"
KYMA_RELEASE=${1:-$defaultRelease}

OVERRIDES_MINIKUBE_PATH="${ROOT_PATH}"/installation/resources/installer-overrides-minikube.yaml
OVERRIDES_COMPASS_GATEWAY="${ROOT_PATH}"/installation/resources/installer-overrides-compass-gateway.yaml
INSTALLER_CR_PATH="${ROOT_PATH}"/installation/resources/installer-cr-kyma-diet-compass.yaml

kyma provision minikube
MINIKUBE_IP=$(eval minikube ip)
sed -i.bak 's/IP_PLACEHOLDER/'$MINIKUBE_IP'/g' $OVERRIDES_MINIKUBE_PATH
kyma install -o $INSTALLER_CR_PATH -o $OVERRIDES_MINIKUBE_PATH -o $OVERRIDES_COMPASS_GATEWAY --release "${KYMA_RELEASE}"
mv -f ${OVERRIDES_MINIKUBE_PATH}.bak $OVERRIDES_MINIKUBE_PATH

# TODO: Remove it after next CLI release
echo "Adding Compass entries to /etc/hosts..."
sudo sh -c 'echo "\n$(minikube ip) compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local" >> /etc/hosts'
