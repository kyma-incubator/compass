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

echo CONNECTOR_URL=${APP_CONNECTOR_URL}
echo SECURED_CONNECTOR_URL=${APP_SECURED_CONNECTOR_URL}

kubectl -n compass-system port-forward svc/compass-connector 3000:3000 &
PORT_FWD_PID=$!

kubectl -n compass-system port-forward svc/compass-connector 8080:8080 &
PORT_FWD_PID_2=$!

export APP_INTERNAL_CONNECTOR_URL=http://localhost:3000/graphql
export APP_HYDRATOR_URL=http://localhost:8080

echo "Wait 5s for port forward to handle requests properly..."
sleep 5

pushd ${CURRENT_DIR}/..

go clean --testcache

go test ./...

popd

kill ${PORT_FWD_PID}
kill ${PORT_FWD_PID_2}
