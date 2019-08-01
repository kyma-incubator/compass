#!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )


POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="11"

DB_USER="usr"
DB_PWD="pwd"
DB_NAME="compass"
DB_PORT="5432"

function cleanup() {
    echo -e "${GREEN}Cleanup Postgres container${NC}"
    docker rm --force ${POSTGRES_CONTAINER}
}

trap cleanup EXIT

echo -e "${GREEN}Start Postgres in detached mode${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PWD} \
            -e POSTGRES_DB=${DB_NAME} \
            -p 5432:5432 \
            postgres:${POSTGRES_VERSION}

set +e
echo '# WAITING FOR CONNECTION WITH DATABASE #'
for i in {1..30}
do
    pg_isready -U "$DB_USER" -h 127.0.0.1 -p "$DB_PORT" -d "$DB_NAME"
    if [ $? -eq 0 ]
    then
        dbReady=true
        break
    fi
    sleep 1
done

if [ "${dbReady}" != true ] ; then
    echo '# COULD NOT ESTABLISH CONNECTION TO DATABASE #'
    exit 1
fi

set -e

echo -e "${GREEN}Populate DB${NC}"

PGPASSWORD=pwd psql -U ${DB_USER} -h 127.0.0.1 -f <(cat ${ROOT_PATH}/../schema-migrator/migrations/*.up.sql) ${DB_NAME}

APP_DB_USER=${DB_USER} APP_DB_PASSWORD=${DB_PWD} APP_DB_NAME=${DB_NAME} go run ${ROOT_PATH}/cmd/main.go