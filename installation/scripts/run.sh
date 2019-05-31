#!/bin/sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

kyma provision minikube

kyma install -o "${ROOT_PATH}"/installation/resources/installer-cr.yaml --release "1.1.0"

#Get Tiller tls client certificates
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
echo "Secrets with Tiller tls client certificates have been created \n"

helm install --name "compass" "${ROOT_PATH}"/chart/compass --tls