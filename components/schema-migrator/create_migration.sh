#!/usr/bin/env bash

DATE="$(date +%Y%m%d%H%M)"
MIGRATIONS_DIR="./migrations/"

touch "${MIGRATIONS_DIR}${DATE}_name.up.sql"
touch "${MIGRATIONS_DIR}${DATE}_name.down.sql"