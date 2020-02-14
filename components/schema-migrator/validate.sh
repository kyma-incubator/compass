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

DB_USER="usr"
DB_PWD="pwd"
DB_PORT="5432"
DB_SSL_PARAM="disable"
POSTGRES_MULTIPLE_DATABASES="compass, broker, provisioner"

function cleanup() {
    echo -e "${GREEN}Cleanup Postgres container and network${NC}"
    docker rm --force ${POSTGRES_CONTAINER}
    docker network rm ${NETWORK}
}

trap cleanup EXIT

echo -e "${GREEN}Create network${NC}"
docker network create --driver bridge ${NETWORK}

docker build -t ${IMG_NAME} ./

echo -e "${GREEN}Start Postgres in detached mode${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
            --network=${NETWORK} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PWD} \
            -e POSTGRES_MULTIPLE_DATABASES="${POSTGRES_MULTIPLE_DATABASES}" \
            -v $(pwd)/multiple-postgresql-databases.sh:/docker-entrypoint-initdb.d/multiple-postgresql-databases.sh \
            postgres:${POSTGRES_VERSION}

function migrationUP() {
    echo -e "${GREEN}Run UP migrations ${NC}"

    migration_path=$1
    db_name=$2
    docker run --rm --network=${NETWORK} \
            -e DB_USER=${DB_USER} \
            -e DB_PASSWORD=${DB_PWD} \
            -e DB_HOST=${POSTGRES_CONTAINER} \
            -e DB_PORT=${DB_PORT} \
            -e DB_NAME=${db_name} \
            -e DB_SSL=${DB_SSL_PARAM} \
            -e MIGRATION_PATH=${migration_path} \
            -e DIRECTION="up" \
        ${IMG_NAME}

    echo -e "${GREEN}Show schema_migrations table after UP migrations${NC}"
    docker exec ${POSTGRES_CONTAINER} psql -U usr compass -c "select * from schema_migrations"
}

function migrationDOWN() {
    echo -e "${GREEN}Run DOWN migrations ${NC}"

    migration_path=$1
    db_name=$2
    docker run --rm --network=${NETWORK} \
            -e DB_USER=${DB_USER} \
            -e DB_PASSWORD=${DB_PWD} \
            -e DB_HOST=${POSTGRES_CONTAINER} \
            -e DB_PORT=${DB_PORT} \
            -e DB_NAME=${db_name} \
            -e DB_SSL=${DB_SSL_PARAM} \
            -e MIGRATION_PATH=${migration_path} \
            -e DIRECTION="down" \
            -e NON_INTERACTIVE="true" \
        ${IMG_NAME}

    echo -e "${GREEN}Show schema_migrations table after DOWN migrations${NC}"
    docker exec ${POSTGRES_CONTAINER} psql -U usr compass -c "select * from schema_migrations"
}

function migrationProcess() {
    path=$1
    db=$2

    echo -e "${GREEN}Migrations for \"${db}\" database and \"${path}\" path${NC}"
    migrationUP "${path}" "${db}"
    migrationDOWN "${path}" "${db}"
}

migrationProcess "director" "compass"
migrationProcess "kyma-environment-broker" "broker"
migrationProcess "provisioner" "provisioner"
