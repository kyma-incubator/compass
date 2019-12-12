#!/bin/bash

set -eu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${CURRENT_DIR}/vars.sh" || { echo 'Cannot load variable file.'; exit 1; }
source "${CURRENT_DIR}/utilities.sh" || { echo 'Cannot load utilities file.'; exit 1; }

for var in CLUSTER_NAME REGION; do
    if [ -z "${!var}" ] ; then
        print_error "ERROR: $var is not set"
        return 1
    fi
done

NETWORK_NAME=${CLUSTER_NAME}-network
ADDRESS_NAME=${CLUSTER_NAME}-nat-ip

print_ok "Create addresses"
gcloud compute addresses create ${ADDRESS_NAME} \
    --region ${REGION}

print_ok "Create routers"
gcloud compute routers create ${CLUSTER_NAME}-nat-router \
    --network ${NETWORK_NAME} \
    --region ${REGION}

print_ok "Create routers nats"
gcloud compute routers nats create ${CLUSTER_NAME}-nat-config \
    --router-region ${REGION} \
    --router ${CLUSTER_NAME}-nat-router \
    --nat-external-ip-pool=${ADDRESS_NAME} \
    --nat-all-subnet-ip-ranges \
