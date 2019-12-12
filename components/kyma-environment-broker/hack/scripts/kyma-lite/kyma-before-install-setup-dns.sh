#!/bin/bash

set -eu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${CURRENT_DIR}/../vars.sh" || { echo 'Cannot load variable file.'; exit 1; }
source "${CURRENT_DIR}/../utilities.sh" || { echo 'Cannot load utilities file.'; exit 1; }

for var in PROJECT DOMAIN_CLUSTER_NAME DNS_NAME CERT_ISSUER_EMAIL CREATE_CERT; do
    if [ -z "${!var}" ] ; then
        print_error "ERROR: $var is not set"
        return 1
    fi
done

# check if docker is running; docker ps -q should only work if the daemon is ready
docker ps -q > /dev/null

DOMAIN="${DOMAIN_CLUSTER_NAME}.${DNS_NAME}"
print_ok "DOMAIN: $DOMAIN"

if [[ "${CREATE_CERT}" == "true" ]]; then
  print_ok "Create cert"
  mkdir -p "${CURRENT_DIR}/letsencrypt"

  SA_NAME="dnsmanager"
  if [[ -z $(gcloud iam service-accounts list | grep $SA_NAME) ]];then
    print_ok "Create service-accounts"
    gcloud iam service-accounts create $SA_NAME --display-name "${SA_NAME}" --project "$PROJECT"
  fi

  print_ok "Add iam-policy"
  gcloud projects add-iam-policy-binding $PROJECT \
      --member serviceAccount:dnsmanager@$PROJECT.iam.gserviceaccount.com --role roles/dns.admin

  print_ok "Create service-accounts keys"
  gcloud iam service-accounts keys create "${CURRENT_DIR}/letsencrypt/key.json" --iam-account dnsmanager@$PROJECT.iam.gserviceaccount.com

  print_ok "Create cert"
  docker run -it --name certbot --rm \
    -v "${CURRENT_DIR}/letsencrypt:/etc/letsencrypt" \
    certbot/dns-google \
    certonly \
    -m $CERT_ISSUER_EMAIL --agree-tos --no-eff-email \
    --dns-google \
    --dns-google-credentials /etc/letsencrypt/key.json \
    --server https://acme-v02.api.letsencrypt.org/directory \
    -d "*.$DOMAIN"
fi

TLS_CERT=$(cat "${CURRENT_DIR}/letsencrypt/live/${DOMAIN}/fullchain.pem" | base64 | sed 's/ /\\ /g' | tr -d '\n');
TLS_KEY=$(cat "${CURRENT_DIR}/letsencrypt/live/${DOMAIN}/privkey.pem" | base64 | sed 's/ /\\ /g' | tr -d '\n')

if [[ -z "${TLS_CERT}" || -z "${TLS_KEY}" ]]; then
  print_error "TLS Cert/Key does not exist"
  return 1
fi

print_ok "Create namespace and configmaps with cert"
kubectl create namespace kyma-installer \
&& kubectl create configmap owndomain-overrides -n kyma-installer --from-literal=global.domainName=$DOMAIN --from-literal=global.tlsCrt=$TLS_CERT --from-literal=global.tlsKey=$TLS_KEY \
&& kubectl label configmap owndomain-overrides -n kyma-installer installer=overrides

kubectl -n kyma-installer get cm owndomain-overrides -oyaml

if [[ -z $(kubectl -n kyma-installer get cm owndomain-overrides -ojson | jq -r '.data."global.tlsCrt"') || -z $(kubectl -n kyma-installer get cm owndomain-overrides -ojson | jq -r '.data."global.tlsKey"') ]];then
  print_error "TLS Cert/Key was not injected into configmap"
  return 1
fi
