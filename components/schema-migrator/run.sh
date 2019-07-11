#!/usr/bin/env bash

for var in DB_USER DB_HOST DB_NAME DB_PORT DB_PASSWORD; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
  exit 1
fi

if [ -n "${DB_SSL}" ] ; then
  DB_NAME="${DB_NAME}?sslmode=${DB_SSL}"
fi

CONNECTION_STRING="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"

migrate -path migrations -database "$CONNECTION_STRING" up
