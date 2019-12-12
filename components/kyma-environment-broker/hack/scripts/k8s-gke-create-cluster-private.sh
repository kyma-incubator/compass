#!/bin/bash

set -eu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${CURRENT_DIR}/vars.sh" || { echo 'Cannot load variable file.'; exit 1; }
source "${CURRENT_DIR}/utilities.sh" || { echo 'Cannot load utilities file.'; exit 1; }

for var in CLUSTER_NAME PROJECT REGION ZONE CLUSTER_NAME CREATE_NETWORK; do
    if [ -z "${!var}" ] ; then
        print_error "ERROR: $var is not set"
        return 1
    fi
done

ADDONS="HorizontalPodAutoscaling,HttpLoadBalancing"
NETWORK_NAME=${CLUSTER_NAME}-network
SUBNET_NAME=${NETWORK_NAME}-192

if [[ "${CREATE_NETWORK}" == "true" ]]; then
  print_ok "Create network"
  gcloud compute networks create ${NETWORK_NAME} \
    --subnet-mode custom

  print_ok "Create network subnets"
  gcloud compute networks subnets create ${SUBNET_NAME} \
   --network ${NETWORK_NAME} \
   --region ${REGION} \
   --range 192.168.20.0/24
fi

print_ok "Create cluster ${CLUSTER_NAME}"
gcloud container --project "${PROJECT}" \
    clusters create "${CLUSTER_NAME}" \
    --zone "${ZONE}" \
    --cluster-version "1.14" \
    --machine-type "n1-standard-4" \
    --addons "${ADDONS}" \
    --enable-private-nodes \
    --enable-ip-alias \
    --network "projects/${PROJECT}/global/networks/${NETWORK_NAME}" \
    --subnetwork "projects/${PROJECT}/regions/${REGION}/subnetworks/${SUBNET_NAME}" \
    --master-ipv4-cidr "172.16.0.0/28" \
    --no-enable-master-authorized-networks \
    --num-nodes 3
