#!/usr/bin/env bash

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

defaultRelease="master"
KYMA_RELEASE=${1:-$defaultRelease}
COMPASS_HELM_RELEASE_NAME="compass"
COMPASS_HELM_RELEASE_NAMESPACE="compass-system"

CR_KYMA_LITE_PATH="${ROOT_PATH}"/installation/resources/installer-cr-kyma-lite.yaml

kyma provision minikube
kyma install -o $CR_KYMA_LITE_PATH --release "${KYMA_RELEASE}"

#Get Tiller tls client certificates
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
echo "Secrets with Tiller tls client certificates have been created \n"

MINIKUBE_IP=$(eval minikube ip)
helm install --set=global.minikubeIP=${MINIKUBE_IP} --set=global.isLocalEnv=true --name "${COMPASS_HELM_RELEASE_NAME}" --namespace "${COMPASS_HELM_RELEASE_NAMESPACE}" "${ROOT_PATH}"/chart/compass --tls --wait
