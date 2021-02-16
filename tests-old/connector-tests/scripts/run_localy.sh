#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

kubectl -n compass-system port-forward svc/compass-connector 3000:3000 &
PORT_FWD_PID=$!

kubectl -n compass-system port-forward svc/compass-connector 3001:3001 &
PORT_FWD_PID_2=$!

kubectl -n compass-system port-forward svc/compass-connector 8080:8080 &
PORT_FWD_PID_3=$!

export APP_CONNECTOR_URL=$(kubectl describe deploy -n compass-system compass-director | grep APP_CONNECTOR_URL | tr -s " " | cut -d " " -f 3)
export APP_EXTERNAL_CONNECTOR_URL=http://localhost:3000/graphql
export APP_INTERNAL_CONNECTOR_URL=http://localhost:3001/graphql
export APP_HYDRATOR_URL=http://localhost:8080

echo "APP_CONNECTOR_URL=$APP_CONNECTOR_URL"
echo "APP_EXTERNAL_CONNECTOR_URL=$APP_EXTERNAL_CONNECTOR_URL"
echo "APP_INTERNAL_CONNECTOR_URL=$APP_INTERNAL_CONNECTOR_URL"
echo "APP_HYDRATOR_URL=$APP_HYDRATOR_URL"

echo "Wait 5s for port forward to handle requests properly..."
sleep 5

pushd ${CURRENT_DIR}/..

go clean --testcache

go test ./...

popd

kill ${PORT_FWD_PID}
kill ${PORT_FWD_PID_2}
kill ${PORT_FWD_PID_3}
