#!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SKIP_DB_CLEANUP=false
REUSE_DB=false
DUMP_DB=false
DISABLE_ASYNC_MODE=true

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --skip-app-start)
            SKIP_APP_START=true
            shift # past argument
        ;;
        --skip-db-cleanup)
            SKIP_DB_CLEANUP=true
            shift
        ;;
        --reuse-db)
            REUSE_DB=true
            shift
        ;;
        --dump-db)
            DUMP_DB=true
            shift
        ;;
        --debug)
            DEBUG=true
            DEBUG_PORT=40000
            shift
        ;;
        --async-enabled)
          DISABLE_ASYNC_MODE=false
          shift
        ;;
        --debug-port)
            DEBUG_PORT=$2
            shift
            shift
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="11"

DB_USER="postgres"
DB_PWD="pgsql@12345"
DB_NAME="compass"
DB_PORT="5432"
DB_HOST="127.0.0.1"

function cleanup() {
    if [[ ${DEBUG} ]]; then
       echo -e "${GREEN}Cleanup Director binary${NC}"
       rm  $GOPATH/src/github.com/kyma-incubator/compass/components/director/director
    fi

    if [[ ${SKIP_DB_CLEANUP} = false ]]; then
        echo -e "${GREEN}Cleanup Postgres container${NC}"
        docker rm --force ${POSTGRES_CONTAINER}
    else
        echo -e "${GREEN}Skipping Postgres container cleanup${NC}"
    fi
}

function cleanup_temp_migrations() {
    rm ${ROOT_PATH}/../schema-migrator/migrations/new_migrations.sql
}

trap cleanup EXIT

if [[ ${REUSE_DB} = true ]]; then
    echo -e "${GREEN}Will reuse existing Postgres container${NC}"
else
    set +e
    echo -e "${GREEN}Start Postgres in detached mode${NC}"
    docker run -d --name ${POSTGRES_CONTAINER} \
                -e POSTGRES_HOST=${DB_HOST} \
                -e POSTGRES_USER=${DB_USER} \
                -e POSTGRES_PASSWORD=${DB_PWD} \
                -e POSTGRES_DB=${DB_NAME} \
                -e POSTGRES_PORT=${DB_PORT} \
                -p ${DB_PORT}:${DB_PORT} \
                postgres:${POSTGRES_VERSION}

    if [[ $? -ne 0 ]] ; then
        SKIP_DB_CLEANUP=true
        exit 1
    fi

    echo '# WAITING FOR CONNECTION WITH DATABASE #'
    for i in {1..30}
    do
        docker exec ${POSTGRES_CONTAINER} pg_isready -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"
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

    if [[ ${DUMP_DB} = false ]]; then
        cat ${ROOT_PATH}/../schema-migrator/migrations/director/*.up.sql | \
            docker exec -i ${POSTGRES_CONTAINER} psql -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"

        cat ${ROOT_PATH}/../schema-migrator/seeds/director/*.sql | \
            docker exec -i ${POSTGRES_CONTAINER} psql -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"
    else
        bash ${ROOT_PATH}/../schema-migrator/dump_db.sh

        cat ${ROOT_PATH}/../schema-migrator/seeds/dump.sql | \
            docker exec -i ${POSTGRES_CONTAINER} psql -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"

        CONNECTION_STRING="postgres://$DB_USER:$DB_PWD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
        migrate -path ${ROOT_PATH}/../schema-migrator/migrations/director -database "$CONNECTION_STRING" up
    fi
fi

INTERNAL_TENANT_ID=$(docker exec -i ${POSTGRES_CONTAINER} psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT id FROM business_tenant_mappings WHERE external_tenant = '3e64ebae-38b5-46a0-b1ed-9ccee153a0ae'")
echo -e "${GREEN}Internal Tenant ID for default tenant from dump: $INTERNAL_TENANT_ID${NC}"

JWT_TOKEN=$(go run ${ROOT_PATH}/cmd/jwtgenerator/main.go -tenant $INTERNAL_TENANT_ID)
echo -e "${GREEN}Use the following JWT token when requesting Director as default tenant:${NC}"
echo $JWT_TOKEN


if [[  ${SKIP_APP_START} ]]; then
    echo -e "${GREEN}Skipping starting application${NC}"
    while true
    do
        sleep 1
    done
fi

echo -e "${GREEN}Starting application${NC}"

export APP_DB_USER=${DB_USER}
export APP_DB_PASSWORD=${DB_PWD}
export APP_DB_NAME=${DB_NAME}
export APP_CONFIGURATION_FILE=${ROOT_PATH}/hack/config-local.yaml
export APP_STATIC_USERS_SRC=${ROOT_PATH}/hack/static-users-local.yaml
export APP_STATIC_GROUPS_SRC=${ROOT_PATH}/hack/static-groups-local.yaml
export APP_OAUTH20_URL="https://oauth2-admin.kyma.local"
export APP_OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT="https://oauth2.kyma.local/oauth2/token"
export APP_ONE_TIME_TOKEN_URL="http://connector.not.configured.url/graphql"
export APP_URL="http://director.not.configured.url/director"
export APP_CONNECTOR_URL="http://connector.not.configured.url/connector/graphql"
export APP_LEGACY_CONNECTOR_URL="https://adapter-gateway.kyma.local/v1/applications/signingRequests/info"
export APP_LOG_LEVEL=debug
export APP_DISABLE_ASYNC_MODE=${DISABLE_ASYNC_MODE}

if [[  ${DEBUG} ]]; then
    echo -e "${GREEN}Debug mode activated on port $DEBUG_PORT${NC}"
    cd $GOPATH/src/github.com/kyma-incubator/compass/components/director
    CGO_ENABLED=0 go build -gcflags="all=-N -l" ./cmd/director
    dlv --listen=:$DEBUG_PORT --headless=true --api-version=2 exec ./director
else
    go run ${ROOT_PATH}/cmd/director/main.go
fi
