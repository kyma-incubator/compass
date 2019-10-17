#!/usr/bin/env bash

set -e
set -o errexit
set -o nounset
set -o pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

NETWORK="gen-examples-network"
POSTGRES_CONTAINER="compass-dev-postgres"
DIRECTOR_CONTAINER="compass-dev-director"
POSTGRES_VERSION="11"
MIGRATOR_IMG_NAME="compass-schema-migrator"
DIRECTOR_IMG_NAME="compass-director"
DB_USER="usr"
DB_PWD="pwd"
DB_NAME="compass"
DB_PORT="5432"
DB_SSL_PARAM="disable"
APP_PORT="3001"

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIRECTOR_URL="compass-dev-director"
LOCAL_ROOT_PATH=${SCRIPT_DIR}/../..

if [ -z ${HOST_ROOT_PATH+x} ];
then
    # create network, because we are run locally
    HOST_ROOT_PATH=${SCRIPT_DIR}/../..
    DIRECTOR_URL="localhost"
    echo -e "${GREEN}Creating network...${NC}"
    docker network create --driver bridge ${NETWORK}
fi;

function cleanup() {
    echo -e "${GREEN}Cleaning up...${NC}"
    docker rm --force ${POSTGRES_CONTAINER} || true
    docker rm --force ${DIRECTOR_CONTAINER} || true
}

trap cleanup EXIT

echo -e "${GREEN}Running database container...${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
    --network=${NETWORK} \
    -e POSTGRES_USER=${DB_USER} \
    -e POSTGRES_PASSWORD=${DB_PWD} \
    -e POSTGRES_DB=${DB_NAME} \
    postgres:${POSTGRES_VERSION}

echo -e "${GREEN}Building migration image...${NC}"
cd "${SCRIPT_DIR}/../../components/schema-migrator/" && docker build -t $MIGRATOR_IMG_NAME ./

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

echo -e "${GREEN}Building Director image...${NC}"

cd "${SCRIPT_DIR}/../../components/director/"
mkdir -p ./licenses
dep ensure --vendor-only -v
docker build -t $DIRECTOR_IMG_NAME ./

echo -e "${GREEN}Running Director...${NC}"

SCOPES_CONFIGURATION_FILE_PATH="${HOST_ROOT_PATH}/chart/compass/charts/director/config.yaml"

docker run --name ${DIRECTOR_CONTAINER} -d --rm --network=${NETWORK} \
    -p ${APP_PORT}:${APP_PORT} \
    -v "${SCOPES_CONFIGURATION_FILE_PATH}:/app/config.yaml" \
    -v "${HOST_ROOT_PATH}/components/director/hack/default-jwks.json:/app/default-jwks.json" \
    -e APP_ADDRESS=0.0.0.0:${APP_PORT} \
    -e APP_DB_USER=${DB_USER} \
    -e APP_DB_PASSWORD=${DB_PWD} \
    -e APP_DB_HOST=${POSTGRES_CONTAINER} \
    -e APP_DB_PORT=${DB_PORT} \
    -e APP_DB_NAME=${DB_NAME} \
    -e APP_SCOPES_CONFIGURATION_FILE=/app/config.yaml \
    -e APP_OAUTH20_CLIENT_ENDPOINT="https://oauth2-admin.kyma.local/clients" \
    -e APP_OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT="https://oauth2.kyma.local/oauth2/token" \
    -e APP_ONE_TIME_TOKEN_URL=http://connector.not.configured.url/graphql \
    -e APP_CONNECTOR_URL=http://connector.not.configured.url/graphql \
    -e APP_JWKS_ENDPOINT="file:///app/default-jwks.json" \
    ${DIRECTOR_IMG_NAME}

cd "${SCRIPT_DIR}"

export DIRECTOR_URL="http://${DIRECTOR_URL}:${APP_PORT}"
./wait-for-director.sh

echo -e "${GREEN}Removing previous GraphQL examples...${NC}"
rm -f "${LOCAL_ROOT_PATH}/examples/"*

echo -e "${GREEN}Running Director tests with generating examples...${NC}"
go test -c "${SCRIPT_DIR}/director/" -tags no_token_test
ALL_SCOPES="runtime:write application:write label_definition:write integration_system:write application:read runtime:read label_definition:read integration_system:read health_checks:read" \
./director.test

echo -e "${GREEN}Prettifying GraphQL examples...${NC}"
img="prettier:latest"
docker build -t ${img} ./tools/prettier
docker run -v "${HOST_ROOT_PATH}/examples":/prettier/examples \
    ${img} prettier --write ./examples/*.graphql

cd "${SCRIPT_DIR}/tools/example-index-generator/"
EXAMPLES_DIRECTORY="${LOCAL_ROOT_PATH}/examples" go run main.go
