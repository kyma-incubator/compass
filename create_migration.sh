#!/usr/bin/env bash

DATE="$(date +%Y%m%d%H%M)"
MIGRATIONS_DIR="./components/schema-migrator/migrations/"

touch "${MIGRATIONS_DIR}${DATE}_name.up.sql"
touch "${MIGRATIONS_DIR}${DATE}_name.down.sql"