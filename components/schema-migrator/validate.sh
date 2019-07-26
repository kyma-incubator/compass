#!/usr/bin/env bash

# This script is responsible for validating if migrations scripts are correct.
# It starts Postgres, executes UP and DOWN migrations.
# This script requires `compass-schema-migrator` Docker image.

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e


IMG_NAME="compass-schema-migrator"
NETWORK="migration-test-network"
POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="11"

function cleanup() {
    echo -e "${GREEN}Cleanup Postgres container and network${NC}"
    docker rm --force ${POSTGRES_CONTAINER}
    docker network rm ${NETWORK}
}

trap cleanup EXIT

echo -e "${GREEN}Create network${NC}"
docker network create --driver bridge ${NETWORK}

echo -e "${GREEN}Start Postgres in detached mode${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
            --network=${NETWORK} \
            -e POSTGRES_USER="usr" \
            -e POSTGRES_PASSWORD="pwd" \
            -e POSTGRES_DB="compass" \
            postgres:${POSTGRES_VERSION}

echo -e "${GREEN}Run UP migrations ${NC}"

docker run --rm --network=${NETWORK} \
        -e DB_USER="usr" \
        -e DB_PASSWORD="pwd" \
        -e DB_HOST=${POSTGRES_CONTAINER} \
        -e DB_PORT="5432" \
        -e DB_NAME="compass" \
        -e DB_SSL="disable" \
        -e DIRECTION="up" \
    ${IMG_NAME}

echo -e "${GREEN}Show schema_migrations table after UP migrations${NC}"
docker exec ${POSTGRES_CONTAINER} psql -U usr compass -c "select * from schema_migrations"

echo -e "${GREEN}Run DOWN migrations ${NC}"
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

echo -e "${GREEN}Show schema_migrations table after DOWN migrations${NC}"
docker exec ${POSTGRES_CONTAINER} psql -U usr compass -c "select * from schema_migrations"