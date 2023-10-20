#!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --debug)
            DEBUG=true
            DEBUG_PORT=40001
            shift
        ;;
        --debug-port)
            DEBUG_PORT=$2
            shift
            shift
        ;;
        --jwks-endpoint)
            export APP_JWKS_ENDPOINT=$2
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

GCLOUD_LOGGED=$(gcloud auth list --format="json" | jq '. | length')

if [[  ${GCLOUD_LOGGED} == "0" ]]; then
    echo -e "${RED}Login to GCloud. Use 'gcloud auth login'. ${NC}" 
    exit 1
fi

POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="12"

DB_USER="postgres"
DB_PWD="pgsql@12345"
DB_NAME="compass"
DB_PORT="5432"
DB_HOST="127.0.0.1"

K3D_CONTEXT="k3d-k3d-cluster"
STAGE_CONTEXT="gke_sap-cp-cmp-stage_europe-west1_sap-cp-cmp-stage"

CLIENT_CERT_SECRET_NAMESPACE="default"
CLIENT_CERT_SECRET_NAME="external-client-certificate"
EXT_SVC_CERT_SECRET_NAME="ext-svc-client-certificate"

function cleanup() {
    if [[ ${DEBUG} == true ]]; then
       echo -e "${GREEN}Cleanup ORD Aggregator ${NC}"
       rm  $GOPATH/src/github.com/kyma-incubator/compass/components/director/ordAggregator
    fi
    rm -fr $GOPATH/src/github.com/kyma-incubator/compass/components/director/run
}

mkdir $GOPATH/src/github.com/kyma-incubator/compass/components/director/run

trap cleanup EXIT

echo -e "${GREEN}Starting application${NC}"

export APP_DB_USER=${DB_USER}
export APP_DB_PASSWORD=${DB_PWD}
export APP_DB_HOST=${DB_HOST}
export APP_DB_PORT=${DB_PORT}
export APP_DB_NAME=${DB_NAME}
#export APP_DIRECTOR_GRAPHQL_URL="http://localhost:3000/graphql"
#export APP_DIRECTOR_SKIP_SSL_VALIDATION="true"
#export APP_DIRECTOR_REQUEST_TIMEOUT="30s"
#export APP_SYSTEM_INFORMATION_PARALLELLISM="1"
#export APP_SYSTEM_INFORMATION_QUEUE_SIZE="1"
#export APP_ENABLE_SYSTEM_DELETION="false"
#export APP_OPERATIONAL_MODE="DISCOVER_SYSTEMS"
#export APP_SYSTEM_INFORMATION_FETCH_TIMEOUT="30s"
#export APP_SYSTEM_INFORMATION_PAGE_SIZE="200"
#export APP_SYSTEM_INFORMATION_PAGE_SKIP_PARAM='$skip'
#export APP_SYSTEM_INFORMATION_PAGE_SIZE_PARAM='$top'
#export APP_OAUTH_TENANT_HEADER_NAME="x-zid"
#export APP_OAUTH_SCOPES_CLAIM="uaa.resource"
#export APP_OAUTH_TOKEN_PATH="/oauth/token"
#export APP_OAUTH_TOKEN_ENDPOINT_PROTOCOL="https"
#export APP_OAUTH_TOKEN_REQUEST_TIMEOUT="30s"
#export APP_OAUTH_SKIP_SSL_VALIDATION="false"
export APP_DB_SSL="disable"
export APP_LOG_FORMAT="json"
export APP_DB_MAX_OPEN_CONNECTIONS="5"
export APP_DB_MAX_IDLE_CONNECTIONS="2"
export APP_SKIP_SSL_VALIDATION="true"
export APP_ADDRESS="0.0.0.0:3001"
export APP_ROOT_API="/ord-aggregator"
export APP_ALLOW_JWT_SIGNING_NONE="true"
export APP_HTTP_RETRY_ATTEMPTS="3"
export APP_HTTP_RETRY_DELAY="100ms"
export APP_EXTERNAL_CLIENT_CERT_SECRET=${CLIENT_CERT_SECRET_NAMESPACE}/${CLIENT_CERT_SECRET_NAME}-stage
export APP_EXTERNAL_CLIENT_CERT_KEY="tls.crt"
export APP_EXTERNAL_CLIENT_KEY_KEY="tls.key"
export APP_EXTERNAL_CLIENT_CERT_SECRET_NAME=${CLIENT_CERT_SECRET_NAME}-stage
export APP_EXT_SVC_CLIENT_CERT_SECRET=${CLIENT_CERT_SECRET_NAMESPACE}/${EXT_SVC_CERT_SECRET_NAME}-stage
export APP_EXT_SVC_CLIENT_CERT_KEY="tls.crt"
export APP_EXT_SVC_CLIENT_KEY_KEY="tls.key"
export APP_EXT_SVC_CLIENT_CERT_SECRET_NAME=${EXT_SVC_CERT_SECRET_NAME}-stage
export APP_ELECTION_LEASE_LOCK_NAME="aggregatorlease"
export APP_ELECTION_LEASE_LOCK_NAMESPACE="default"
export APP_OPERATIONS_MANAGER_PRIORITY_QUEUE_LIMIT="10"
export APP_OPERATIONS_MANAGER_RESCHEDULE_JOB_INTERVAL="24h"
export APP_OPERATIONS_MANAGER_RESCHEDULE_PERIOD="168h"
export APP_OPERATIONS_MANAGER_RESCHEDULE_HANGED_JOB_INTERVAL="1h"
export APP_OPERATIONS_MANAGER_RESCHEDULE_HANGED_PERIOD="1h"
export APP_MAINTAIN_OPERATIONS_JOB_INTERVAL="60m"
export APP_OPERATION_PROCESSORS_QUIET_PERIOD="5s"
export APP_PARALLEL_OPERATION_PROCESSORS="10"
export APP_MAX_PARALLEL_DOCUMENTS_PER_APPLICATION="10"
export APP_MAX_PARALLEL_SPECIFICATION_PROCESSORS="100"
export APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY="subscriptionProviderId"
export APP_ORD_WEBHOOK_MAPPINGS="[]"
export APP_TENANT_MAPPING_CALLBACK_URL="http://tenant-mapping.not.configured.url/director"
export APP_TENANT_MAPPING_CONFIG_PATH="$GOPATH/src/github.com/kyma-incubator/compass/components/director/run/tntMappingConfig.json"
export APP_CREDENTIAL_EXCHANGE_STRATEGY_TENANT_MAPPINGS='{"ASYNC_CALLBACK": {}}'
export APP_CONFIGURATION_FILE="$GOPATH/src/github.com/kyma-incubator/compass/components/director/run/config.yaml"
export HOSTNAME="local"

echo '{"ASYNC_CALLBACK": {}}' >> $GOPATH/src/github.com/kyma-incubator/compass/components/director/run/tntMappingConfig.json

# Fetch needed artifacts from stage cluster
kubectl config use-context ${STAGE_CONTEXT}
#kubectl get configmap compass-system-fetcher-templates-config -n compass-system -o json | jq -r '.data."app-templates.json"' | jq -r '.' > ${APP_TEMPLATES_FILE_LOCATION}/app-templates.json
kubectl get configmap compass-director-config -n compass-system -o json | jq -r '.data."config.yaml"' > ${APP_CONFIGURATION_FILE}
#export APP_OAUTH_CLIENT_ID=$(kubectl get secret xsuaa-instance -n compass-system -o json | jq -r '.data."x509.credentials.clientid"' | base64 --decode)
#export APP_OAUTH_TOKEN_BASE_URL=$(kubectl get secret xsuaa-instance -n compass-system -o json | jq -r '.data."x509.credentials.certurl"' | base64 --decode)
export APP_EXTERNAL_CLIENT_CERT_VALUE=$(kubectl get secret -n compass-system ${CLIENT_CERT_SECRET_NAME} -o json | jq -r '.data."tls.crt"' | base64 --decode)
export APP_EXTERNAL_CLIENT_KEY_VALUE=$(kubectl get secret -n compass-system ${CLIENT_CERT_SECRET_NAME} -o json | jq -r '.data."tls.key"' | base64 --decode)
export APP_EXT_SVC_CLIENT_CERT_VALUE=$(kubectl get secret -n compass-system ${EXT_SVC_CERT_SECRET_NAME} -o json | jq -r '.data."tls.crt"' | base64 --decode)
export APP_EXT_SVC_CLIENT_KEY_VALUE=$(kubectl get secret -n compass-system ${EXT_SVC_CERT_SECRET_NAME} -o json | jq -r '.data."tls.key"' | base64 --decode)

ENV_VARS=$(kubectl get pod -n compass-system $(kubectl get pods -n compass-system  | grep "ord-aggregator" | awk '{print $1}' | head -1) -o=jsonpath='{.spec.containers[?(@.name=="ord-aggregator")]}' | jq -r '.env')

# Adjust artifacts inside local cluster
kubectl config use-context ${K3D_CONTEXT}
kubectl create secret generic "$CLIENT_CERT_SECRET_NAME"-stage --from-literal="$APP_EXTERNAL_CLIENT_CERT_KEY"="$APP_EXTERNAL_CLIENT_CERT_VALUE" --from-literal="$APP_EXTERNAL_CLIENT_KEY_KEY"="$APP_EXTERNAL_CLIENT_KEY_VALUE" --save-config --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic "$EXT_SVC_CERT_SECRET_NAME"-stage --from-literal="$APP_EXT_SVC_CLIENT_CERT_KEY"="$APP_EXT_SVC_CLIENT_CERT_VALUE" --from-literal="$APP_EXT_SVC_CLIENT_KEY_KEY"="$APP_EXT_SVC_CLIENT_KEY_VALUE" --save-config --dry-run=client -o yaml | kubectl apply -f -

export APP_GLOBAL_REGISTRY_URL="$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_GLOBAL_REGISTRY_URL") | .value' )"

. ${ROOT_PATH}/hack/jwt_generator.sh

# Start Debug or Run mode
if [[  ${DEBUG} == true ]]; then
    echo -e "${GREEN}Debug mode activated on port $DEBUG_PORT${NC}"
    cd $GOPATH/src/github.com/kyma-incubator/compass/components/director
    CGO_ENABLED=0 go build -gcflags="all=-N -l" ./cmd/ordaggregator
    dlv --listen=:$DEBUG_PORT --headless=true --api-version=2 exec ./ordaggregator
else
    go run ${ROOT_PATH}/cmd/ordaggregator/main.go
fi
