ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
POSTGRES_CONTAINER="test-postgres-replica"
POSTGRES_VERSION="15"

DB_USER="postgres"
DB_PWD="pgsql@12345"
DB_NAME="compass"
DB_PORT="6432"
DB_HOST="127.0.0.1"

docker run -d --name ${POSTGRES_CONTAINER} \
                -e POSTGRES_HOST=${DB_HOST} \
                -e POSTGRES_USER=${DB_USER} \
                -e POSTGRES_PASSWORD=${DB_PWD} \
                -e POSTGRES_DB=${DB_NAME} \
                -e POSTGRES_PORT=${DB_PORT} \
                -p ${DB_PORT}:${DB_PORT} \
                -v ${ROOT_PATH}/../schema-migrator/seeds:/tmp \
                postgres:${POSTGRES_VERSION}