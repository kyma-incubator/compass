#!/bin/bash

###
# Following script uninstalls compass installation only.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --sql-helm-backend)
            SQL_HELM_BACKEND=true
            shift # past argument
        ;;
    esac
done

if [[ ${SQL_HELM_BACKEND} ]]; then
    echo -e "${GREEN}Helm SQL storage backend will be used${NC}"

    DB_USER=$(base64 -d <<< $(kubectl get secret -n compass-system compass-postgresql -o=jsonpath="{.data['postgresql-director-username']}"))
    DB_PWD=$(base64 -d <<< $(kubectl get secret -n compass-system compass-postgresql -o=jsonpath="{.data['postgresql-director-password']}"))
    DB_NAME=$(base64 -d <<< $(kubectl get secret -n compass-system compass-postgresql -o=jsonpath="{.data['postgresql-directorDatabaseName']}"))
    DB_PORT=$(base64 -d <<< $(kubectl get secret -n compass-system compass-postgresql -o=jsonpath="{.data['postgresql-servicePort']}"))

    kubectl port-forward --namespace compass-system svc/compass-postgresql ${DB_PORT}:${DB_PORT} &
    sleep 5 #wait port-forwarding to be completed

    export HELM_DRIVER=sql
    export HELM_DRIVER_SQL_CONNECTION_STRING=postgres://${DB_USER}:${DB_PWD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable
fi

echo "Wait for helm stable status"
wait_for_helm_stable_state "compass" "compass-system" 

echo "Uninstall Compass"
helm uninstall --wait --debug --timeout "${TIMEOUT}" --namespace compass-system compass || true

if [[ ${SQL_HELM_BACKEND} ]]; then
    pkill kubectl
fi