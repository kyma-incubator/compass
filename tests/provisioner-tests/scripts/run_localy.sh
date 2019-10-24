#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

for var in  APP_GCP_CREDENTIALS APP_GCP_PROJECT_NAME; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

kubectl -n compass-system port-forward svc/compass-provisioner 3000:3000 &
PORT_FWD_PID=$!

export APP_INTERNAL_PROVISIONER_URL=http://localhost:3000/graphql

echo "Wait 5s for port forward to handle requests properly..."
sleep 5

pushd ${CURRENT_DIR}/..

go clean --testcache

go test ./...

popd

kill ${PORT_FWD_PID}
