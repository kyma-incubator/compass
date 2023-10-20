#!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
ORD_DATA_CREATION=true
ORD_DATA_CLEANUP=true
APP_TEMPLATE_NAME="SAP local ord aggregator template"
APP_NAME="ord-test-app"
APP_TEMPLATE_ID=""
APP_ID=""

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --skip-data-cleanup)
            ORD_DATA_CLEANUP=false
            shift
        ;;
        --skip-data-creation)
            ORD_DATA_CREATION=false
            shift
        ;;
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
    rm -fr ${ROOT_PATH}/../../open-resource-discovery-reference-application

    if [[ ${ORD_DATA_CLEANUP} == true ]]; then
        if [[ -z $APP_ID ]]; then
          echo -e "${RED} APP_ID variable is null ${NC}"
          return 1
        fi

        if [[ -z $APP_TEMPLATE_ID ]]; then
          echo -e "${RED} APP_TEMPLATE_ID variable is null ${NC}"
          return 1
        fi

        . ${ROOT_PATH}/hack/jwt_generator.sh
        DIRECTOR_TOKEN="$(get_token | tr -d '\n')"

        echo -e "${YELLOW} Cleaning up Application with ID ${APP_ID} ${NC}"
        DELETE_APP="mutation { result: unregisterApplication(id:\"${APP_ID}\") { id } }"
        execute_gql_query "${APP_DIRECTOR_GRAPHQL_URL}" "${DIRECTOR_TOKEN}" "${DELETE_APP}"

        echo -e "\n${YELLOW} Cleaning up Application Template with ID ${APP_TEMPLATE_ID} ${NC}"
        DELETE_APP_TEMPLATE="mutation { result: deleteApplicationTemplate(id:\"${APP_TEMPLATE_ID}\") { id }"
        execute_gql_query "${APP_DIRECTOR_GRAPHQL_URL}" "${DIRECTOR_TOKEN}" "${DELETE_APP_TEMPLATE}"

    fi
}

function setup() {
    mkdir $GOPATH/src/github.com/kyma-incubator/compass/components/director/run

    cd ${ROOT_PATH}/../../
    git clone https://github.tools.sap/CentralEngineering/open-resource-discovery-reference-application
    cd open-resource-discovery-reference-application
    npm set @sap:registry=http://nexus.wdf.sap.corp:8081/nexus/content/groups/build.milestones.npm/

    cat > ./src/config.ts << EOF
export const port = process.env.PORT || 8080
export const localUrl = \`http://localhost:\${port}\`
export const publicUrl = \`http://localhost:\${port}\`
EOF

    npm install
    npm run build
    npm start > /dev/null &

    cd ${ROOT_PATH}
}

function execute_gql_query(){
    local URL=${1}
    local DIRECTOR_TOKEN=${2}
    local MUTATION=${3:-""}

    if [ "" != "${MUTATION}" ]; then
        local GQL_QUERY='{ "query": "'${MUTATION}'" }'
    fi
    curl --request POST --url "${URL}" --header "Content-Type: application/json" --header "authorization: Bearer ${DIRECTOR_TOKEN}" -d "${GQL_QUERY}"
}

trap cleanup EXIT

setup

echo -e "${GREEN}Starting application${NC}"

export APP_DIRECTOR_GRAPHQL_URL="http://localhost:3000/graphql"
export APP_DB_USER=${DB_USER}
export APP_DB_PASSWORD=${DB_PWD}
export APP_DB_HOST=${DB_HOST}
export APP_DB_PORT=${DB_PORT}
export APP_DB_NAME=${DB_NAME}
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
kubectl get configmap compass-director-config -n compass-system -o json | jq -r '.data."config.yaml"' > ${APP_CONFIGURATION_FILE}
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

# Create Application Template and Application with ORD webhook if needed
if [[  ${ORD_DATA_CREATION} == true ]]; then
    echo -e "${GREEN}Creating tenant${NC}"
    DIRECTOR_TOKEN="$(get_token | tr -d '\n')"

    CREATE_APP_TEMPLATE_MUTATION="mutation { result: createApplicationTemplate( in: { name: \\\"${APP_TEMPLATE_NAME}\\\" description: \\\"app-template-desc\\\" applicationInput: { name: \\\"{{name}}\\\", providerName: \\\"compass-tests\\\", description: \\\"test {{name}}\\\", webhooks: [{ type: OPEN_RESOURCE_DISCOVERY, url: \\\"http://localhost:8080/.well-known/open-resource-discovery\\\" }] healthCheckURL: \\\"http://url.valid\\\" } placeholders: [ { name: \\\"name\\\" description: \\\"app\\\" jsonPath: \\\"new-placeholder-name-json-path\\\" } ] accessLevel: GLOBAL } ) { id } }"
    CREATE_APP_TEMPLATE_RESULT="$(execute_gql_query "${APP_DIRECTOR_GRAPHQL_URL}" "${DIRECTOR_TOKEN}" "${CREATE_APP_TEMPLATE_MUTATION}")"
    echo -e "${GREEN}Application Template created: ${NC}"
    echo "${CREATE_APP_TEMPLATE_RESULT}"

    APP_TEMPLATE_ID=$(echo "${CREATE_APP_TEMPLATE_RESULT}" | jq '.data.result.id')

    CREATE_APP_MUTATION="mutation { result: registerApplicationFromTemplate( in: { templateName: \\\"${APP_TEMPLATE_NAME}\\\" values: [ { placeholder: \\\"name\\\", value: \\\"${APP_NAME}\\\" } ] } ) { id } }"
    CREATE_APP_RESULT="$(execute_gql_query "${APP_DIRECTOR_GRAPHQL_URL}" "${DIRECTOR_TOKEN}" "${CREATE_APP_MUTATION}")"
    echo -e "${GREEN}Application created: ${NC}"
    echo "${CREATE_APP_RESULT}"

    APP_ID=$(echo "${CREATE_APP_RESULT}" | jq '.data.result.id')
else
    echo -e "${GREEN}Tenant creation skipped${NC}"
fi

echo ${APP_TEMPLATE_ID}
echo ${APP_ID}

# Start Debug or Run mode
if [[  ${DEBUG} == true ]]; then
    echo -e "${GREEN}Debug mode activated on port $DEBUG_PORT${NC}"
    cd $GOPATH/src/github.com/kyma-incubator/compass/components/director
    CGO_ENABLED=0 go build -gcflags="all=-N -l" ./cmd/ordaggregator
    dlv --listen=:$DEBUG_PORT --headless=true --api-version=2 exec ./ordaggregator
else
    go run ${ROOT_PATH}/cmd/ordaggregator/main.go
fi
