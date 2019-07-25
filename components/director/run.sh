#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color


NETWORK="migration-dev-network"
POSTGRES_CONTAINER="compass-dev-postgres"
POSTGRES_VERSION="11"
MIGRATOR_IMG_NAME="compass-schema-migrator"

DB_USER="usr"
DB_PWD="pwd"
DB_NAME="compass"
DB_PORT="5432"
DB_SSL_PARAM="disable"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd ${DIR}

function cleanup() {
	echo -e "${GREEN}Removing database container...${NC}"
    docker rm --force ${POSTGRES_CONTAINER}
    docker network rm ${NETWORK}
}

trap cleanup EXIT

echo -e "${GREEN}Creating network...${NC}"
docker network create --driver bridge ${NETWORK}

echo -e "${GREEN}Running database container...${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
			--network=${NETWORK} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PWD} \
            -e POSTGRES_DB=${DB_NAME} \
            -p 5432:5432 \
            postgres:${POSTGRES_VERSION}

echo -e "${GREEN}Building migration image...${NC}"
cd ../schema-migrator/ && make build-image

echo -e "${GREEN}Running migration...${NC}"

docker run --rm --network=${NETWORK} \
        -e DB_USER=${DB_USER} \
        -e DB_PASSWORD=${DB_PWD} \
        -e DB_HOST=${POSTGRES_CONTAINER} \
        -e DB_PORT=${DB_PORT} \
        -e DB_NAME=${DB_NAME} \
        -e DB_SSL=${DB_SSL_PARAM} \
        -e DIRECTION="up" \
    ${MIGRATOR_IMG_NAME}

echo -e "${GREEN}Running application...${NC}"
cd ../director/
APP_DB_USER=${DB_USER} APP_DB_PASSWORD=${DB_PWD} APP_DB_NAME=${DB_NAME} go run cmd/main.go