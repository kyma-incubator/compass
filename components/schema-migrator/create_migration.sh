#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

COMPONENT=$1
NAME=$2

for var in COMPONENT NAME; do
    if [ -z "${!var}" ] ; then
        echo "One or more arguments not provided. Usage: ./create_migration [COMPONENT] [NAME]"
        exit 1
    fi
done

DATE="$(date +%Y%m%d%H%M)"
MIGRATIONS_DIR="${DIR}/migrations"

touch "${MIGRATIONS_DIR}/${COMPONENT}/${DATE}_${NAME}.up.sql"
touch "${MIGRATIONS_DIR}/${COMPONENT}/${DATE}_${NAME}.down.sql"
