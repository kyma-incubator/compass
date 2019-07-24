#!/usr/bin/env bash

IMG_NAME="compass-schema-migrator"
NETWORK="migration-test-network"
POSTGRES_CONTAINER="test-postgres"

echo "Everything is fine"

function cleanup() {
    echo "cleanup"
  #  docker rm --force ${POSTGRES_CONTAINER}
   # docker network rm ${NETWORK}
}

trap cleanup EXIT

docker network create --driver bridge ${NETWORK}


docker run -d --name ${POSTGRES_CONTAINER} \
            --network=${NETWORK} \
            -p 5432:5432 \
            -e POSTGRES_USER="usr" \
            -e POSTGRES_PASSWORD="pwd" \
            -e POSTGRES_DB="compass" \
            postgres:11

sleep 5

docker run --rm --network=${NETWORK} \
        -e DB_USER="usr" \
        -e DB_PASSWORD="pwd" \
        -e DB_HOST=${POSTGRES_CONTAINER} \
        -e DB_PORT="5432" \
        -e DB_NAME="compass" \
        -e DB_SSL="disable" \
        -e DIRECTION="up" \
    ${IMG_NAME}


docker run --rm --network=${NETWORK} \
        -e DB_USER="usr" \
        -e DB_PASSWORD="pwd" \
        -e DB_HOST=${POSTGRES_CONTAINER} \
        -e DB_PORT="5432" \
        -e DB_NAME="compass" \
        -e DB_SSL="disable" \
        -e DIRECTION="down" \
        -e NON_INTERACTIVE="true" \
    ${IMG_NAME}

docker exec ${POSTGRES_CONTAINER} psql