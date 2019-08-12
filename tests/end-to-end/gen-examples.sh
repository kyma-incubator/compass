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
COMPASS_PROJECT_PATH="${GOPATH}/src/github.com/kyma-incubator/compass"
DB_USER="usr"
DB_PWD="pwd"
DB_NAME="compass"
DB_PORT="5432"
DB_SSL_PARAM="disable"
APP_PORT="3001"

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function cleanup {
	echo -e "${GREEN}Cleaning up...${NC}"
    docker rm --force ${POSTGRES_CONTAINER} || true
    docker rm --force ${DIRECTOR_CONTAINER} || true
    docker network rm ${NETWORK} || true
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
            postgres:${POSTGRES_VERSION}

echo -e "${GREEN}Building migration image...${NC}"
cd "${SCRIPT_DIR}/../../components/schema-migrator/" && make build-image

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

cd "${SCRIPT_DIR}/../../components/director/" && make resolve build-image

echo -e "${GREEN}Running Director...${NC}"

docker run -d --name ${DIRECTOR_CONTAINER} --rm --network=${NETWORK} \
		-p ${APP_PORT}:${APP_PORT} \
		-e APP_ADDRESS=0.0.0.0:${APP_PORT} \
        -e APP_DB_USER=${DB_USER} \
        -e APP_DB_PASSWORD=${DB_PWD} \
        -e APP_DB_HOST=${POSTGRES_CONTAINER} \
        -e APP_DB_PORT=${DB_PORT} \
        -e APP_DB_NAME=${DB_NAME} \
    ${DIRECTOR_IMG_NAME}


cd "${SCRIPT_DIR}"

# wait for Director to be up and running

echo -e "${GREEN}Checking if Director is up...${NC}"
directorIsUp=false
set +e
for i in {1..10}; do
    curl --fail  "http://localhost:${APP_PORT}/graphql"  -H "Content-Type: application/json"  -H 'tenant: any' --data-binary '{"query":"{\n  __schema {\n    queryType {\n      name\n    }\n  }\n}"}'
    res=$?

    if [[ ${res} = 0 ]]
	then
	    directorIsUp=true
	    break
	fi
    sleep 1
done
set -e

echo ""

if [[ "$directorIsUp" = false ]]; then
	echo -e "${RED}Cannot access Director API${NC}"
    exit 1
fi

echo -e "${GREEN}Removing previous GraphQL examples...${NC}"
rm -f "${COMPASS_PROJECT_PATH}/examples/"*

echo -e "${GREEN}Running Director tests with generating examples...${NC}"
go test -c "${SCRIPT_DIR}/director/"
DIRECTOR_GRAPHQL_API="http://localhost:${APP_PORT}/graphql" ./director.test

echo -e "${GREEN}Prettifying GraphQL examples...${NC}"
img="prettier:latest"
docker build -t ${img} ./tools/prettier
docker run -v "${COMPASS_PROJECT_PATH}/examples":/prettier/examples \
            ${img} prettier --write ./examples/*.graphql

cd "${SCRIPT_DIR}/tools/example-index-generator/"
EXAMPLES_DIRECTORY="${COMPASS_PROJECT_PATH}/examples" go run main.go
