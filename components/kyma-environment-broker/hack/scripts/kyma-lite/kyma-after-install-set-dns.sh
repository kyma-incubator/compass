#!/usr/bin/env bash

set -eu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${CURRENT_DIR}/../vars.sh" || { echo 'Cannot load variable file.'; exit 1; }
source "${CURRENT_DIR}/../utilities.sh" || { echo 'Cannot load utilities file.'; exit 1; }

for var in PROJECT DOMAIN_CLUSTER_NAME DNS_NAME DNS_ZONE; do
    if [ -z "${!var}" ] ; then
        print_error "ERROR: $var is not set"
        return 1
    fi
done

DOMAIN="${DOMAIN_CLUSTER_NAME}.${DNS_NAME}"

export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

if [[ -z "${EXTERNAL_PUBLIC_IP}" || -z "${APISERVER_PUBLIC_IP}" ]]; then
  print_error "External public IP/Apiserver public IP does not exist"
  return 1
fi

print_ok "Add dns records for IP: ${EXTERNAL_PUBLIC_IP}, ${APISERVER_PUBLIC_IP}"
gcloud dns --project=$PROJECT record-sets transaction start --zone=$DNS_ZONE
gcloud dns --project=$PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE
gcloud dns --project=$PROJECT record-sets transaction add $APISERVER_PUBLIC_IP --name=\apiserver.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE
gcloud dns --project=$PROJECT record-sets transaction execute --zone=$DNS_ZONE
