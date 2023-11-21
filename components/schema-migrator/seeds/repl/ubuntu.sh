#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install postgresql -y
apt-get install postgresql-14-wal2json -y

echo "test-postgres-replica:5432:compass:postgres:pgsql@12345" > /.pgpass
chmod 0600 ./.pgpass
export PGPASSFILE='/.pgpass'

echo "Sleep while databases get set-up"
sleep 40

pg_recvlogical -h test-postgres-replica -d compass -U postgres --slot test_slot --create-slot -P wal2json
pg_recvlogical -h test-postgres-replica -d compass  -U postgres --slot test_slot --start -o pretty-print=1 -o add-msg-prefixes=wal2json -f -