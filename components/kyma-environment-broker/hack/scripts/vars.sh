#!/usr/bin/env bash

readonly CLUSTER_NAME="gophers-cis-whitelisted"

readonly PROJECT="sap-se-cx-gopher"
readonly REGION="europe-west1"
readonly ZONE="europe-west1-d"

readonly DOMAIN_CLUSTER_NAME="cis"
readonly DNS_NAME="gophers.kyma.pro"
readonly DNS_ZONE="gophers-kyma-pro"
readonly CERT_ISSUER_EMAIL="adam.walach@sap.com"

# specifies create new tls certificate or not (if not use existing cert)
readonly CREATE_CERT="false"
# specifies create new VPC network (not create if network already exist)
readonly CREATE_NETWORK="false"
