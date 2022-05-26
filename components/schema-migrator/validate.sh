#!/usr/bin/env bash

# This script is responsible for validating if migrations scripts are correct.
# It starts Postgres, executes UP and DOWN migrations.
# This script requires `compass-schema-migrator` Docker image.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

COMPONENT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

IMG_NAME="compass-schema-migrator"
NETWORK="migration-test-network"
POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="14"

PROJECT="sap-cp-cmp"
ENV="dev"

DB_USER="usr"
DB_PWD="pwd"
DB_PORT="5432"
DB_SSL_PARAM="disable"
POSTGRES_MULTIPLE_DATABASES="compass"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --dump-db)
            DUMP_DB=true
            shift # past argument
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

function cleanup() {
    echo -e "${GREEN}Cleanup Postgres container and network${NC}"
    docker rm --force ${POSTGRES_CONTAINER}
    docker network rm ${NETWORK}
}

trap cleanup EXIT

if [[ ${DUMP_DB} ]] ; then
    echo -e "${GREEN}DB dump will be used to validate migrations${NC}"
    if [[ ! -f ${COMPONENT_PATH}/seeds/dump.sql ]]; then
        echo -e "${YELLOW}Will pull DB dump from GCR bucket${NC}"
        gsutil cp gs://$PROJECT-$ENV-db-dump/dump.sql "${COMPONENT_PATH}"/seeds/dump.sql
    else
        echo -e "${GREEN}DB dump already exists on system, will reuse it${NC}"
    fi
else
    rm -f ${COMPONENT_PATH}/seeds/dump.sql # necessary, so that a dump is not left behind by accident and used in the building of the image
fi

echo -e "${GREEN}Create network${NC}"
docker network create --driver bridge ${NETWORK}

docker build -t ${IMG_NAME} ./

echo -e "${GREEN}Start Postgres in detached mode${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
            --network=${NETWORK} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PWD} \
            -e POSTGRES_MULTIPLE_DATABASES="${POSTGRES_MULTIPLE_DATABASES}" \
            -p 5432:5432 \
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
            -e DRY_RUN="true" \
        ${IMG_NAME}

    echo -e "${GREEN}Show schema_migrations table after UP migrations${NC}"
    docker exec ${POSTGRES_CONTAINER} psql --username usr "${db_name}" --command "select * from schema_migrations"
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
            -e DRY_RUN="true" \
        ${IMG_NAME}

    echo -e "${GREEN}Show schema_migrations table after DOWN migrations${NC}"
    docker exec ${POSTGRES_CONTAINER} psql --username usr "${db_name}" --command "select * from schema_migrations"
}

function migrationProcess() {
    path=$1
    db=$2

    echo -e "${GREEN}Migrations for \"${db}\" database and \"${path}\" path${NC}"
    migrationUP "${path}" "${db}"

    if [[ ! -f seeds/dump.sql ]]; then
        migrationDOWN "${path}" "${db}"
    fi
}

migrationProcess "director" "compass"
