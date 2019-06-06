#!/bin/sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

KYMA_RELEASE="1.1.0"
COMPASS_HELM_RELEASE_NAME="compass"
COMPASS_HELM_RELEASE_NAMESPACE="compass-system"

kyma provision minikube

kyma install -o "${ROOT_PATH}"/installation/resources/installer-cr.yaml --release "${KYMA_RELEASE}"

#Get Tiller tls client certificates
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
echo "Secrets with Tiller tls client certificates have been created \n"

helm install --name "${COMPASS_HELM_RELEASE_NAME}" --namespace "${COMPASS_HELM_RELEASE_NAMESPACE}" "${ROOT_PATH}"/chart/compass --tls

echo "Adding Compass entries to /etc/hosts...\n"
sudo sh -c 'echo "\n$(minikube ip) compass-gateway.kyma.local" >> /etc/hosts'