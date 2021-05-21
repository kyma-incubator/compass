#!/usr/bin/env bash

for var in DB_USER DB_HOST DB_NAME DB_PORT DB_PASSWORD MIGRATION_PATH DIRECTION; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

if [[ "${DIRECTION}" == "up" ]]; then
    echo "Migration UP"
elif [[ "${DIRECTION}" == "down" ]]; then
    echo "Migration DOWN"
else
    echo "ERROR: DIRECTION variable accepts only two values: up or down"
    exit 1
fi

echo '# WAITING FOR CONNECTION WITH DATABASE #'
for i in {1..30}
do
    pg_isready -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME"
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

DB_NAME_SSL=$DB_NAME
SSL_OPTION=""
if [ -n "${DB_SSL}" ] ; then
  DB_NAME_SSL="${DB_NAME}?sslmode=${DB_SSL}"
  SSL_OPTION="sslmode=${DB_SSL}"
fi


if [[ -f seeds/dump.sql ]]; then
    echo "Will reuse existing dump in seeds/dump.sql"
    cat seeds/dump.sql | \
        PGPASSWORD="${DB_PASSWORD}" psql -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" --set="${SSL_OPTION}"

    REMOTE_MIGRATION_VERSION=$(PGPASSWORD="${DB_PASSWORD}" psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT version FROM schema_migrations")
    LOCAL_MIGRATION_VERSION=$(echo $(ls migrations/director | tail -n 1) | grep -o -E '[0-9]+' | head -1 | sed -e 's/^0\+//')

    if [[ ${REMOTE_MIGRATION_VERSION} = ${LOCAL_MIGRATION_VERSION} ]]; then
        echo "Both remote and local migrations are at the same version."
    else
        echo "NOTE: Remote and local migrations are at different versions."
        echo "REMOTE: $REMOTE_MIGRATION_VERSION"
        echo "LOCAL: $LOCAL_MIGRATION_VERSION"
    fi
fi

CONNECTION_STRING="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME_SSL"

CMD="migrate -path migrations/${MIGRATION_PATH} -database "$CONNECTION_STRING" ${DIRECTION}"
echo '# STARTING MIGRATION #'
if [[ "${NON_INTERACTIVE}" == "true" ]]; then
    yes | $CMD
else
    $CMD
fi
