#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..
defaultRelease=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)
KYMA_RELEASE=${1:-$defaultRelease}
COMPASS_HELM_RELEASE_NAME="compass"
COMPASS_HELM_RELEASE_NAMESPACE="compass-system"
INSTALLER_CR_PATH="${ROOT_PATH}"/installation/resources/installer-cr-kyma-diet.yaml
OVERRIDES_COMPASS_GATEWAY="${ROOT_PATH}"/installation/resources/installer-overrides-compass-gateway.yaml
ISTIO_OVERRIDES="${ROOT_PATH}"/installation/resources/installer-overrides-istio.yaml
API_GATEWAY_OVERRIDES="${ROOT_PATH}"/installation/resources/installer-overrides-api-gateway.yaml
MINIKUBE_HELM_VALUES="${ROOT_PATH}"/installation/resources/minikube-values.yaml

kyma provision minikube
kyma install -o $INSTALLER_CR_PATH  -o $OVERRIDES_COMPASS_GATEWAY -o $ISTIO_OVERRIDES -o $API_GATEWAY_OVERRIDES "${KYMA_RELEASE}"

#Get Tiller tls client certificates
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
echo -e "Secrets with Tiller tls client certificates have been created \n"

bash "${ROOT_PATH}"/installation/scripts/run-compass-installer.sh
bash "${ROOT_PATH}"/installation/scripts/is-installed.sh

# TODO: Remove it after next CLI release
echo "Adding Compass entries to /etc/hosts..."
sudo sh -c 'echo "\n$(minikube ip) adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local" >> /etc/hosts'
