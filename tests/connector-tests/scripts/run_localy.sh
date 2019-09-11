#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

for var in APP_CONNECTOR_URL APP_SECURED_CONNECTOR_URL; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

kubectl -n compass-system port-forward svc/compass-connector 3000:3000 &
PORT_FWD_PID=$!

export APP_INTERNAL_CONNECTOR_URL=http://localhost:3000/graphql

pushd ${CURRENT_DIR}/..

go test ./...

popd

kill ${PORT_FWD_PID}
