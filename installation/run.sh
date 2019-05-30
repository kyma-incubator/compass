#!/bin/sh

ROOT_DIR=$(dirname ${BASH_SOURCE})/..

kyma provision minikube

kyma install -o ${ROOT_DIR}/installation/installer-cr.yaml

${ROOT_DIR}/installation/tiller-tls.sh

helm install ${ROOT_DIR}/chart/compass --tls