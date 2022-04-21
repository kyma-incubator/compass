#!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.
set -e

. "func.source"

die_on_noval ${GOPATH} "GOPATH is mandatory"

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
log_info ${ROOT_PATH}

SKIP_DB_CLEANUP=false
REUSE_DB=false
DUMP_DB=false
DISABLE_ASYNC_MODE=true
DEBUG=false
DEBUG_PORT=40000
COMPONENT='director'

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"
    case ${key} in
        --skip-app-start)
            SKIP_APP_START=true
            shift # past argument
        ;;
        --skip-db-cleanup)
            SKIP_DB_CLEANUP=true
            shift
        ;;
        --reuse-db)
            REUSE_DB=true
            shift
        ;;
        --dump-db)
            DUMP_DB=true
            shift
        ;;
        --debug)
            DEBUG=true
            shift
        ;;
        --async-enabled)
          DISABLE_ASYNC_MODE=false
          shift
        ;;
        --tenant-fetcher)
          COMPONENT='tenantfetcher-svc'
          shift
        ;;
        --ns-adapter)
          COMPONENT='ns-adapter'
          export APP_SYSTEM_TO_TEMPLATE_MAPPINGS='[{  "Name": "S4HANA",  "SourceKey": ["type"],  "SourceValue": ["on-premise"]}]'
          shift
        ;;
        --jwks-endpoint)
          export APP_JWKS_ENDPOINT=$2
          shift
          shift
        ;;
        --debug-port)
            DEBUG_PORT=$2
            shift
            shift
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

trap "cleanup ${DEBUG} ${SKIP_DB_CLEANUP} ${POSTGRES_CONTAINER}" EXIT

log_info "Creating k3d cluster..."
install_k3d
create_k3d_cluster

log_info "Initialise DB ..."
create_db ${ROOT_PATH} ${REUSE_DB} ${DUMP_DB}

read -r CURRENT_MIGRATION_VERSION <<< $(get_applied_migration_version)
log_info "Migration version: ${CURRENT_MIGRATION_VERSION}"

log_info "Token:"
log_info "------- start -------"
get_token
log_info "------- end -------"

if [[  ${SKIP_APP_START} ]]; then
    log_info "Skipping starting application"
    while true
    do
        sleep 1
    done
fi

log_info "Starting application"

start_app ${ROOT_PATH} ${DISABLE_ASYNC_MODE} ${DEBUG} ${DEBUG_PORT} ${COMPONENT}