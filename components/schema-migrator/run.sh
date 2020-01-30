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

if [ -n "${DB_SSL}" ] ; then
  DB_NAME="${DB_NAME}?sslmode=${DB_SSL}"
fi

CONNECTION_STRING="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"

CMD="migrate -path migrations/${MIGRATION_PATH} -database "$CONNECTION_STRING" ${DIRECTION}"
echo '# STARTING MIGRATION #'
if [[ "${NON_INTERACTIVE}" == "true" ]]; then
    yes | $CMD
else
    $CMD
fi
