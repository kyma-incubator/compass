git checkout #!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
TENANT_CREATION=true
APP_VERIFY_TENANT=""

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --skip-tenant-creation)
            TENANT_CREATION=false
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
        --tenant)
            APP_VERIFY_TENANT=$2
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

# Exit when tenant is not provided
if [[  ${APP_VERIFY_TENANT} == "" ]]; then
    echo -e "${RED}Tenant not provided. Use --tenant. ${NC}" 
    exit 1
fi

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
       echo -e "${GREEN}Cleanup System Fetcher ${NC}"
       rm  $GOPATH/src/github.com/kyma-incubator/compass/components/director/systemfetcher
    fi
    rm -fr $GOPATH/src/github.com/kyma-incubator/compass/components/director/run
}


trap cleanup EXIT

function execute_gql_query(){
    local URL=${1}
    local DIRECTOR_TOKEN=${2}
    local MUTATION=${3:-""}

    if [ "" != "${MUTATION}" ]; then
        local GQL_QUERY='{ "query": "'${MUTATION}'" }'
    fi
    curl --request POST --url "${URL}" --header "Content-Type: application/json" --header "authorization: Bearer ${DIRECTOR_TOKEN}" -d "${GQL_QUERY}" 
}

echo -e "${GREEN}Starting application${NC}"

export APP_DB_USER=${DB_USER}
export APP_DB_PASSWORD=${DB_PWD}
export APP_DB_HOST=${DB_HOST}
export APP_DB_PORT=${DB_PORT}
export APP_DB_NAME=${DB_NAME}
export APP_DIRECTOR_GRAPHQL_URL="http://localhost:3000/graphql"
export APP_DIRECTOR_SKIP_SSL_VALIDATION="true"
export APP_DIRECTOR_REQUEST_TIMEOUT="30s"
export APP_SYSTEM_INFORMATION_PARALLELLISM="1"
export APP_SYSTEM_INFORMATION_QUEUE_SIZE="1"
export APP_ENABLE_SYSTEM_DELETION="false"
export APP_OPERATIONAL_MODE="DISCOVER_SYSTEMS"
export APP_SYSTEM_INFORMATION_FETCH_TIMEOUT="30s"
export APP_SYSTEM_INFORMATION_PAGE_SIZE="200"
export APP_SYSTEM_INFORMATION_PAGE_SKIP_PARAM='$skip'
export APP_SYSTEM_INFORMATION_PAGE_SIZE_PARAM='$top'
export APP_OAUTH_TENANT_HEADER_NAME="x-zid"
export APP_OAUTH_SCOPES_CLAIM="uaa.resource"
export APP_OAUTH_TOKEN_PATH="/oauth/token"
export APP_OAUTH_TOKEN_ENDPOINT_PROTOCOL="https"
export APP_OAUTH_TOKEN_REQUEST_TIMEOUT="30s"
export APP_OAUTH_SKIP_SSL_VALIDATION="false"
export APP_DB_SSL="disable"
export APP_LOG_FORMAT="kibana"
export APP_DB_MAX_OPEN_CONNECTIONS="5"
export APP_DB_MAX_IDLE_CONNECTIONS="2"
export APP_EXTERNAL_CLIENT_CERT_SECRET=${CLIENT_CERT_SECRET_NAMESPACE}/${CLIENT_CERT_SECRET_NAME}-stage
export APP_EXTERNAL_CLIENT_CERT_KEY="tls.crt"
export APP_EXTERNAL_CLIENT_KEY_KEY="tls.key"
export APP_EXTERNAL_CLIENT_CERT_SECRET_NAME=${CLIENT_CERT_SECRET_NAME}-stage
export APP_EXT_SVC_CLIENT_CERT_SECRET=${CLIENT_CERT_SECRET_NAMESPACE}/${EXT_SVC_CERT_SECRET_NAME}-stage
export APP_EXT_SVC_CLIENT_CERT_KEY="tls.crt"
export APP_EXT_SVC_CLIENT_KEY_KEY="tls.key"
export APP_EXT_SVC_CLIENT_CERT_SECRET_NAME=${EXT_SVC_CERT_SECRET_NAME}-stage
export APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY="xsappname"
export APP_CONFIGURATION_FILE="$GOPATH/src/github.com/kyma-incubator/compass/components/director/run/config.yaml"
export APP_TEMPLATES_FILE_LOCATION="$GOPATH/src/github.com/kyma-incubator/compass/components/director/run/templates/"

mkdir -p ${APP_TEMPLATES_FILE_LOCATION} || true

# Fetch needed artifacts from stage cluster
kubectl config use-context ${STAGE_CONTEXT}
kubectl get configmap compass-system-fetcher-templates-config -n compass-system -o json | jq -r '.data."app-templates.json"' | jq -r '.' > ${APP_TEMPLATES_FILE_LOCATION}/app-templates.json
kubectl get configmap compass-director-config -n compass-system -o json | jq -r '.data."config.yaml"' > ${APP_CONFIGURATION_FILE}
export APP_OAUTH_CLIENT_ID=$(kubectl get secret xsuaa-instance -n compass-system -o json | jq -r '.data."x509.credentials.clientid"' | base64 --decode)
export APP_OAUTH_TOKEN_BASE_URL=$(kubectl get secret xsuaa-instance -n compass-system -o json | jq -r '.data."x509.credentials.certurl"' | base64 --decode)
export APP_EXTERNAL_CLIENT_CERT_VALUE=$(kubectl get secret -n compass-system ${CLIENT_CERT_SECRET_NAME} -o json | jq -r '.data."tls.crt"' | base64 --decode)
export APP_EXTERNAL_CLIENT_KEY_VALUE=$(kubectl get secret -n compass-system ${CLIENT_CERT_SECRET_NAME} -o json | jq -r '.data."tls.key"' | base64 --decode)
export APP_EXT_SVC_CLIENT_CERT_VALUE=$(kubectl get secret -n compass-system ${EXT_SVC_CERT_SECRET_NAME} -o json | jq -r '.data."tls.crt"' | base64 --decode)
export APP_EXT_SVC_CLIENT_KEY_VALUE=$(kubectl get secret -n compass-system ${EXT_SVC_CERT_SECRET_NAME} -o json | jq -r '.data."tls.key"' | base64 --decode)

ENV_VARS=$(kubectl get cronjob -n compass-system compass-system-fetcher -o=jsonpath='{.spec.jobTemplate.spec.template.spec.containers[?(@.name=="system-fetcher")]}' | jq -r '.env')

export APP_SYSTEM_INFORMATION_ENDPOINT=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_SYSTEM_INFORMATION_ENDPOINT") | .value' )
export APP_SYSTEM_INFORMATION_FILTER_CRITERIA=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_SYSTEM_INFORMATION_FILTER_CRITERIA") | .value')
export APP_SYSTEM_INFORMATION_SOURCE_KEY=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_SYSTEM_INFORMATION_SOURCE_KEY") | .value')
export APP_TEMPLATE_LABEL_FILTER=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_TEMPLATE_LABEL_FILTER") | .value')
export APP_TEMPLATE_OVERRIDE_APPLICATION_INPUT=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_TEMPLATE_OVERRIDE_APPLICATION_INPUT") | .value')
export APP_TEMPLATE_PLACEHOLDER_TO_SYSTEM_KEY_MAPPINGS=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_TEMPLATE_PLACEHOLDER_TO_SYSTEM_KEY_MAPPINGS") | .value' )
export APP_ORD_WEBHOOK_MAPPINGS=$(echo -E ${ENV_VARS} | jq -r '.[] | select(.name == "APP_ORD_WEBHOOK_MAPPINGS") | .value' )

# Adjust artifacts inside local cluster
kubectl config use-context ${K3D_CONTEXT}
kubectl create secret generic "$CLIENT_CERT_SECRET_NAME"-stage --from-literal="$APP_EXTERNAL_CLIENT_CERT_KEY"="$APP_EXTERNAL_CLIENT_CERT_VALUE" --from-literal="$APP_EXTERNAL_CLIENT_KEY_KEY"="$APP_EXTERNAL_CLIENT_KEY_VALUE" --save-config --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic "$EXT_SVC_CERT_SECRET_NAME"-stage --from-literal="$APP_EXT_SVC_CLIENT_CERT_KEY"="$APP_EXT_SVC_CLIENT_CERT_VALUE" --from-literal="$APP_EXT_SVC_CLIENT_KEY_KEY"="$APP_EXT_SVC_CLIENT_KEY_VALUE" --save-config --dry-run=client -o yaml | kubectl apply -f -

# Create tenant if requested
if [[  ${TENANT_CREATION} == true ]]; then
    echo -e "${GREEN}Creating tenant${NC}"
    . ${ROOT_PATH}/hack/jwt_generator.sh
    /Users/i028667/SAPDevelop/go/src/github.com/kyma-incubator/compass/components/director/hack/jwt_generator.sh
    DIRECTOR_TOKEN="$(get_token | tr -d '\n')"

    CREATE_TENANT_MUTATION="mutation { writeTenant(in: { name: \\\"Validation Tenant\\\", externalTenant: \\\"${APP_VERIFY_TENANT}\\\", type: \\\"account\\\", provider: \\\"Compass Tests\\\" })}"
    CREATE_TENANT_RESULT="$(execute_gql_query "${APP_DIRECTOR_GRAPHQL_URL}" "${DIRECTOR_TOKEN}" "${CREATE_TENANT_MUTATION}")"
    echo -e "${GREEN}Tenant created:${NC}"
    echo ${CREATE_TENANT_RESULT}
else
    echo -e "${GREEN}Teant creation skipped${NC}"
fi

# Start Debug or Run mode
if [[  ${DEBUG} == true ]]; then
    echo -e "${GREEN}Debug mode activated on port $DEBUG_PORT${NC}"
    cd $GOPATH/src/github.com/kyma-incubator/compass/components/director
    CGO_ENABLED=0 go build -gcflags="all=-N -l" ./cmd/systemfetcher
    dlv --listen=:$DEBUG_PORT --headless=true --api-version=2 exec ./systemfetcher
else
    go run ${ROOT_PATH}/cmd/systemfetcher/main.go
fi
