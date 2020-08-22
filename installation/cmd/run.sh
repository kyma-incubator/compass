#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

#echo "Provisioning Minikube cluster..."
#kyma provision minikube


#LOCAL_ENV=true bash "${ROOT_PATH}"/installation/scripts/install-minimal-kyma.sh ${1}

bash "${ROOT_PATH}"/installation/scripts/run-compass-installer.sh
bash "${ROOT_PATH}"/installation/scripts/is-installed.sh

echo "Adding Compass entries to /etc/hosts..."
sudo sh -c 'echo "\n$(minikube ip) adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local" >> /etc/hosts'
