#!/usr/bin/env bash

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

defaultRelease="master"
KYMA_RELEASE=${1:-$defaultRelease}

OVERRIDES_MINIKUBE_PATH="${ROOT_PATH}"/installation/resources/installer-overrides-minikube.yaml
CR_KYMA_LITE_COMPASS_PATH="${ROOT_PATH}"/installation/resources/installer-cr-kyma-lite-compass.yaml

kyma provision minikube
MINIKUBE_IP=$(eval minikube ip)
sed -i.bak 's/IP_PLACEHOLDER/'$MINIKUBE_IP'/g' $OVERRIDES_MINIKUBE_PATH
kyma install -o $CR_KYMA_LITE_COMPASS_PATH -o $OVERRIDES_MINIKUBE_PATH --release "${KYMA_RELEASE}"
sed -i.bak 's/'$MINIKUBE_IP'/IP_PLACEHOLDER/g' $OVERRIDES_MINIKUBE_PATH
rm ${OVERRIDES_MINIKUBE_PATH}.bak

# TODO: Remove it after next CLI release
echo "Adding Compass entries to /etc/hosts...\n"
sudo sh -c 'echo "\n$(minikube ip) compass-gateway.kyma.local" >> /etc/hosts'
