#!/bin/bash

docker-entrypoint.sh postgres &
pid_to_wait=$!
sleep 15
echo "slept well"

psql -U "${POSTGRES_USER}" -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -d "${POSTGRES_DB}" -f $1 #/tmp/repl/init-primary.sql
for file in /tmp/migrations/director/*.up.sql; do
  psql -U "${POSTGRES_USER}" -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -d "${POSTGRES_DB}" -f "${file}"
done

for file in /tmp/director/*.sql; do
  psql -U "${POSTGRES_USER}" -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -d "${POSTGRES_DB}" -f "${file}"
done

wait $pid_to_wait