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
CHARTS_PATH=$(dirname $(dirname $COMPONENT_PATH))/chart/compass
CHART_FILE=$CHARTS_PATH/"values.yaml"

DATA_DIR="${COMPONENT_PATH}/seeds"

IMG_NAME="compass-schema-migrator"
CONTAINER_REGISTRY_KEY="containerRegistry"
NETWORK="migration-test-network"
POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="12"

CONTAINER_REGISTRY=$(grep $CONTAINER_REGISTRY_KEY $CHART_FILE -A 1 -m 1 | tail -n 1 | rev | cut -d' ' -f 1 | rev)
IMAGE_VERSION=$(grep $IMG_NAME $CHART_FILE -B 1 | head -n 1 | cut -d'"' -f2)
IMAGE_PULL_LOCATION=$CONTAINER_REGISTRY/$IMG_NAME:$IMAGE_VERSION


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
        --build-image)
            IMAGE_PULL_LOCATION=$IMG_NAME
            BUILD_IMAGE=true
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
    if [[ ${DUMP_DB} ]]; then
      rm -rf "${DATA_DIR}"/dump || true
    fi

    docker network rm ${NETWORK}
}

trap cleanup EXIT

echo -e "${GREEN}Create network${NC}"
docker network create --driver bridge ${NETWORK}

if [[ ${BUILD_IMAGE} ]]; then
  echo -e "${GREEN}Building schema migrator image from Dockerfile${NC}"
  ARCH="amd64"

  if [[ $(uname -m) == 'arm64' ]]; then
      ARCH="arm64"
  fi
  docker build -t ${IMAGE_PULL_LOCATION} ./
else
  echo -e "${GREEN}Pulling schema migrator image from ${IMAGE_PULL_LOCATION}${NC}"
  docker pull "${IMAGE_PULL_LOCATION}"
fi

echo -e "${GREEN}Start Postgres in detached mode${NC}"
docker run -d --name ${POSTGRES_CONTAINER} \
            --network=${NETWORK} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PWD} \
            -e POSTGRES_MULTIPLE_DATABASES="${POSTGRES_MULTIPLE_DATABASES}" \
            -p 5432:5432 \
            -v $(pwd)/multiple-postgresql-databases.sh:/docker-entrypoint-initdb.d/multiple-postgresql-databases.sh \
            -v ${DATA_DIR}:/tmp \
            postgres:${POSTGRES_VERSION}

if [[ ${DUMP_DB} ]]; then
    echo -e "${GREEN}DB dump will be used to prepopulate installation${NC}"

    REMOTE_VERSIONS=($(gsutil ls -R gs://sap-cp-cmp-dev-db-dump/ | grep -o -E '[0-9]+' | sed -e 's/^0\+//' | sort -r))
    LOCAL_VERSIONS=($(ls "$COMPONENT_PATH"/migrations/director | grep -o -E '^[0-9]+' | sed -e 's/^0\+//' | sort -ru))

    SCHEMA_VERSION=""
    for r in "${REMOTE_VERSIONS[@]}"; do
       for l in "${LOCAL_VERSIONS[@]}"; do
          if [[ "$r" == "$l" ]]; then
            SCHEMA_VERSION=$r
            break 2;
          fi
       done
    done

    if [[ -z $SCHEMA_VERSION ]]; then
      echo -e "${RED}\$SCHEMA_VERSION variable cannot be empty${NC}"
    fi

    echo -e "${YELLOW}Check if there is DB dump in GCS bucket with migration number: $SCHEMA_VERSION...${NC}"
    gsutil -q stat gs://sap-cp-cmp-dev-db-dump/dump-"${SCHEMA_VERSION}"/toc.dat
    STATUS=$?

    if [[ $STATUS ]]; then
      echo -e "${GREEN}DB dump with migration number: $SCHEMA_VERSION exists in the bucket. Will use it...${NC}"
    else
      echo -e "${RED}There is no DB dump with migration number: $SCHEMA_VERSION in the bucket.${NC}"
      exit 1
    fi

    if [[ ! -d ${DATA_DIR}/dump-${SCHEMA_VERSION} ]]; then
        echo -e "${YELLOW}There is no dump with number: $SCHEMA_VERSION locally. Will pull the DB dump from GCR bucket...${NC}"
        mkdir ${DATA_DIR}/dump-${SCHEMA_VERSION}
        gsutil cp -r gs://sap-cp-cmp-dev-db-dump/dump-"${SCHEMA_VERSION}" "${DATA_DIR}"
    else
        echo -e "${GREEN}DB dump already exists on the local system, will reuse it${NC}"
    fi
    rm -rf "${DATA_DIR}"/dump || true
    cp -R "${DATA_DIR}"/dump-"${SCHEMA_VERSION}" "${DATA_DIR}"/dump

    echo -e "${GREEN}Starting DB restore process...${NC}"
    docker exec -i ${POSTGRES_CONTAINER} pg_restore --verbose --format=directory --jobs=8 --no-owner --no-privileges --username="${DB_USER}" --host="${DB_HOST}" --port="${DB_PORT}" --dbname="${DB_NAME}" tmp/dump

fi

function migrationUP() {
    echo -e "${GREEN}Run UP migrations using ${IMAGE_PULL_LOCATION}${NC}"

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
        "${IMAGE_PULL_LOCATION}"

    echo -e "${GREEN}Show schema_migrations table after UP migrations${NC}"
    docker exec ${POSTGRES_CONTAINER} psql --username usr "${db_name}" --command "select * from schema_migrations"
}

function migrationDOWN() {
    echo -e "${GREEN}Run DOWN migrations using ${IMAGE_PULL_LOCATION}${NC}"

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
        "${IMAGE_PULL_LOCATION}"

    echo -e "${GREEN}Show schema_migrations table after DOWN migrations${NC}"
    docker exec ${POSTGRES_CONTAINER} psql --username usr "${db_name}" --command "select * from schema_migrations"
}

function migrationProcess() {
    path=$1
    db=$2

    echo -e "${GREEN}Migrations for \"${db}\" database and \"${path}\" path${NC}"
    migrationUP "${path}" "${db}"

    if [[ ! -f "${DATA_DIR}"/dump ]]; then
        migrationDOWN "${path}" "${db}"
    fi
}

migrationProcess "director" "compass"
