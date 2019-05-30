#!/bin/sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

kyma provision minikube

kyma install -o "${ROOT_PATH}"/installation/resources/installer-cr.yaml

"${ROOT_PATH}"/installation/scripts/kyma-scripts/tiller-tls.sh

helm install --name "compass" "${ROOT_PATH}"/chart/compass --tls