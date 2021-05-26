#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

set -e

if [[ -f ${ROOT_PATH}/seeds/dump.sql ]]; then
    echo -e "${GREEN}Will reuse existing DB dump in schema-migrator/seeds/dump.sql ${NC}"
else
    echo -e "${YELLOW}Warning: Please ensure that your kubectl context points to an existing Compass development cluster${NC}"
    sleep 2

    CURRENT_KUBE_CONTEXT=$(kubectl config current-context)
    if [[ $CURRENT_KUBE_CONTEXT != *dev* ]]; then
        echo -e "${RED}Error: Current kubectl context does not point to an existing Compass development cluster${NC}"
        exit 1
    fi

    echo -e "${GREEN}Will dump and use database data from connected kubernetes Compass development cluster${NC}"


    REMOTE_DB_PWD=$(base64 -d <<< $(kubectl get secret -n compass-system compass-postgresql -o=jsonpath="{.data['postgresql-director-password']}"))
    DIRECTOR_POD_NAME=$(kubectl get pods -n compass-system | grep "director" | head -1 | cut -d ' ' -f 1)
    kubectl port-forward --namespace compass-system $DIRECTOR_POD_NAME 5555:5432 &
    sleep 5 # necessary for the port-forward to open in time for the next command

    echo -e "${GREEN}Dumping database. This will take about 2-3 minutes...${NC}"

    PGPASSWORD=$REMOTE_DB_PWD pg_dump --dbname=director --file=${ROOT_PATH}/seeds/dump.sql --host=localhost --port=5555 --username=director --column-inserts --no-owner --no-privileges

    echo -e "${GREEN}Database dumped!${NC}"

    pkill kubectl

fi

set +e
