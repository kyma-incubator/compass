#!/bin/bash

apt-get update
apt-get install curl -y
curl https://techsupport.enterprisedb.com/api/repository/dl/default/release/deb | bash
apt-get install $2

docker-entrypoint.sh postgres &
pid_to_wait=$!

echo "Sleeping 10 sec for postgresql to start-up"
sleep 10

# Check if the PostgreSQL data directory exists
if [ ! -d "$PGDATA" ]; then
  echo "Error: PostgreSQL data directory does not exist"
  exit 1
fi

# Add the configuration to the postgresql.conf file
echo "wal_level = 'logical'" >> "$PGDATA/postgresql.conf"
echo "max_worker_processes = 10" >> "$PGDATA/postgresql.conf"
echo "max_replication_slots = 10" >> "$PGDATA/postgresql.conf"
echo "max_wal_senders = 10" >> "$PGDATA/postgresql.conf"
echo "shared_preload_libraries = 'pglogical'" >> "$PGDATA/postgresql.conf"

# Restart postgresql
kill -9 $pid_to_wait
wait $pid_to_wait

docker-entrypoint.sh postgres &
pid_to_wait=$!
echo "Sleeping 10 sec for postgresql to start-up 2"
sleep 10

psql -U "${POSTGRES_USER}" -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -d "${POSTGRES_DB}" -f $1
for file in /tmp/migrations/director/*.up.sql; do
  psql -U "${POSTGRES_USER}" -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -d "${POSTGRES_DB}" -f "${file}"
done

for file in /tmp/director/*.sql; do
  psql -U "${POSTGRES_USER}" -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -d "${POSTGRES_DB}" -f "${file}"
done

wait $pid_to_wait