#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

kubectl -n compass-system port-forward svc/compass-director 3000:3000 &
PORT_FWD_PID=$!

export APP_DIRECTOR_URL=http://localhost:3000/graphql
export APP_DIRECTOR_HEALTHZ_URL=http://localhost:3000/healthz

echo "APP_DIRECTOR_URL=$APP_DIRECTOR_URL"
echo "APP_DIRECTOR_HEALTHZ_URL=$APP_DIRECTOR_HEALTHZ_URL"

echo "Wait 5s for port forward to handle requests properly..."
sleep 5

pushd ${CURRENT_DIR}/..

echo "Clean up test cache.."
go clean --testcache

echo "Run tests.."
go test ./...

popd

echo "Clean up .."
kill ${PORT_FWD_PID}
