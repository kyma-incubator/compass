#!/bin/bash

###
# Following script installs necessary tooling for Debian, starts Director and Ord service, and runs the smoke tests.
#

set -o errexit

function is_ready(){
    local URL=${1}
    local HTTP_CODE=$(curl -s -o /dev/null -I -w "%{http_code}" ${URL})
    if [ "${HTTP_CODE}" == "200" ]; then
        return 0
    fi 
    echo "Response from ${URL} is still: ${HTTP_CODE}"
    return 1
}

function execute_gql_query(){
    local URL=${1}
    local DIRECTOR_TOKEN=${2}
    local INTERNAL_TENANT_ID=${3}
    local FILE_LOCATION=${4:-""}

    if [ "" != "${FILE_LOCATION}" ]; then
        local FLAT_FILE_CONTENT=$(sed 's/\\/\\\\/g' ${FILE_LOCATION} | sed 's/\"/\\"/g' | sed 's/$/\\n/' | tr -d '\n')
        local GQL_QUERY='{ "query": "'${FLAT_FILE_CONTENT}'" }'
    fi
    curl --request POST --url "${URL}" --header "Content-Type: application/json" --header "authorization: Bearer ${DIRECTOR_TOKEN}" --header "tenant: ${INTERNAL_TENANT_ID}" ${FILE_LOCATION:+"--data"} ${FILE_LOCATION:+"${GQL_QUERY}"} 
}

compare_values() {
    local VAR1=${1}
    local VAR2=${2}
    local MESSAGE=${3}
    if [ "${VAR1}" != "${VAR2}" ]; then
        echo "COMPARE ERROR: ${MESSAGE}"
        TEST_RESULT=false
    fi
}

check_value() {
    local VAR=${1}
    local MESSAGE=${2}
    if [[ "null" == "${VAR}" ]] || [[ -z "${VAR}" ]]; then
        echo "VALIDATION ERROR: ${MESSAGE}"
        TEST_RESULT=false
    fi
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR="$( cd "$( dirname "${CURRENT_DIR}/../../.." )" && pwd )"
BASE_DIR="$( cd "$( dirname "${INSTALLATION_DIR}/../../../../../../.." )" && pwd )"
JAVA_HOME="${BASE_DIR}/openjdk-11"
M2_HOME="${BASE_DIR}/maven"
MIGRATE_HOME="${BASE_DIR}/migrate"
COMPASS_DIR="$( cd "$( dirname "${INSTALLATION_DIR}/../.." )" && pwd )"  
ORD_SVC_DIR="${BASE_DIR}/ord-service"
TEST_RESULT=true

mkdir -p "${JAVA_HOME}"
export JAVA_HOME
export PATH="${JAVA_HOME}/bin:${PATH}"

mkdir -p "${M2_HOME}"
export M2_HOME
export PATH="${M2_HOME}/bin:${PATH}"

mkdir -p "${MIGRATE_HOME}"
export MIGRATE_HOME
export PATH="${MIGRATE_HOME}:${PATH}"

export ARTIFACTS="/var/log/prow_artifacts"
mkdir -p "${ARTIFACTS}"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo "Install java"
curl -fLSs -o adoptopenjdk11.tgz "https://github.com/AdoptOpenJDK/openjdk11-binaries/releases/download/jdk-11.0.9.1%2B1/OpenJDK11U-jdk_x64_linux_hotspot_11.0.9.1_1.tar.gz"
tar --extract --file adoptopenjdk11.tgz --directory "${JAVA_HOME}" --strip-components 1 --no-same-owner
rm adoptopenjdk11.tgz* 

echo "Install maven"
curl -fLSs -o apache-maven-3.8.5.tgz "https://dlcdn.apache.org/maven/maven-3/3.8.5/binaries/apache-maven-3.8.5-bin.tar.gz"
tar --extract --file apache-maven-3.8.5.tgz --directory "${M2_HOME}" --strip-components 1 --no-same-owner
rm apache-maven-3.8.5.tgz* 

echo "Install migrate"
curl -fLSs -o migrate.tgz "https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz"
tar --extract --file migrate.tgz --directory "${MIGRATE_HOME}" --no-same-owner
rm migrate.tgz* 

echo "-----------------------------------"
echo "Eenvironment"
echo "-----------------------------------"
echo "JAVA_HOME: ${JAVA_HOME}"
echo "M2_HOME: ${M2_HOME}"
echo "MIGRATE_HOME: ${MIGRATE_HOME}"
echo "GOPATH: ${GOPATH}"
echo "Base folder: ${BASE_DIR}"
echo "Compass folder: ${COMPASS_DIR}"
echo "ORD-service folder: ${ORD_SVC_DIR}"
echo "Artifacts folder: ${ARTIFACTS}"
echo "Path: ${PATH}"
echo "-----------------------------------"
echo "Java version:"
echo "-----------------------------------"
java -version
echo "-----------------------------------"
echo "Migrate version:"
echo "-----------------------------------"
migrate --version
echo "-----------------------------------"
echo "Go version:"
echo "-----------------------------------"
go version
echo "-----------------------------------"

echo "Starting compass"
cd "${COMPASS_DIR}/components/director"
source run.sh --auto-terminate 300 & 

COMPASS_URL="http://localhost:3000"

START_TIME=$(date +%s)
until is_ready "${COMPASS_URL}/healthz" ; do
    CURRENT_TIME=$(date +%s)
    SECONDS=$((CURRENT_TIME-START_TIME))
    if (( SECONDS > 300 )); then
        echo "Timeout of 5 min for starting compass reached. Exiting."
        exit 1
    fi
    echo "Wait 10s before check Director health again ..."
    sleep 10
done

. ${COMPASS_DIR}/components/director/hack/jwt_generator.sh

DIRECTOR_TOKEN="$(get_token | tr -d '\n')"
INTERNAL_TENANT_ID="$(get_internal_tenant | tr -d '\n')"
echo "Compass is ready"

echo "Starting ord-service"
cd "${ORD_SVC_DIR}/components/ord-service"
export SERVER_PORT=8081
./run.sh --migrations-path ${COMPASS_DIR}/components/schema-migrator/migrations/director --auto-terminate 120 &

ORD_URL="http://localhost:${SERVER_PORT}"

START_TIME=$(date +%s)
until is_ready "${ORD_URL}/actuator/health" ; do
    CURRENT_TIME=$(date +%s)
    SECONDS=$((CURRENT_TIME-START_TIME))
    if (( SECONDS > 300 )); then
        echo "Timeout of 5 min for starting ord-service reached. Exiting."
        exit 1
    fi
    echo "Wait 10s before check Ord Service health again ..."
    sleep 10
done

echo "ORD-service is ready"

echo "Token: ${DIRECTOR_TOKEN}"
echo "Internal Tenant ID: ${INTERNAL_TENANT_ID}"

echo "ord-test start!"

echo "Register applicaiton ..."
REG_APP_FILE_LOCATION="${COMPASS_DIR}/components/director/examples/register-application/register-application-with-bundles.graphql"
COMPASS_GQL_URL="${COMPASS_URL}/graphql"
CREATE_APP_IN_COMPASS_RESULT="$(execute_gql_query "${COMPASS_GQL_URL}" "${DIRECTOR_TOKEN}" "${INTERNAL_TENANT_ID}" "${REG_APP_FILE_LOCATION}")"

echo "Result from app creation request:"
echo "---------------------------------"
echo "${CREATE_APP_IN_COMPASS_RESULT}"
echo "---------------------------------"

CRT_APP_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.id')
CRT_BUNDLES_COUNT=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data | length')

CRT_BUNDLE_0_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].id')
CRT_BUNDLE_0_API_DEFS_COUNT=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].apiDefinitions.data | length')
CRT_BUNDLE_0_API_DEF_0_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].apiDefinitions.data[0].id')
CRT_BUNDLE_0_API_DEF_1_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].apiDefinitions.data[1].id')
CRT_BUNDLE_0_API_DEF_2_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].apiDefinitions.data[2].id')
CRT_BUNDLE_0_EVENT_DEFS_COUNT=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].eventDefinitions.data | length')
CRT_BUNDLE_0_EVENT_DEF_0_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].eventDefinitions.data[0].id')
CRT_BUNDLE_0_EVENT_DEF_1_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[0].eventDefinitions.data[1].id')

CRT_BUNDLE_1_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].id')
CRT_BUNDLE_1_API_DEFS_COUNT=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].apiDefinitions.data | length')
CRT_BUNDLE_1_API_DEF_0_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].apiDefinitions.data[0].id')
CRT_BUNDLE_1_API_DEF_1_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].apiDefinitions.data[1].id')
CRT_BUNDLE_1_API_DEF_2_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].apiDefinitions.data[2].id')
CRT_BUNDLE_1_EVENT_DEFS_COUNT=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].eventDefinitions.data | length')
CRT_BUNDLE_1_EVENT_DEF_0_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].eventDefinitions.data[0].id')
CRT_BUNDLE_1_EVENT_DEF_1_ID=$(echo -E ${CREATE_APP_IN_COMPASS_RESULT} | jq -r '.data.result.bundles.data[1].eventDefinitions.data[1].id')

GET_APPS_FILE_LOCATION="${COMPASS_DIR}/components/director/examples/query-applications/query-applications.graphql"
GET_APPS_FROM_COMPASS_RESULT="$(execute_gql_query "${COMPASS_GQL_URL}" "${DIRECTOR_TOKEN}" "${INTERNAL_TENANT_ID}" "${GET_APPS_FILE_LOCATION}")"

echo "Result from get apps request:"
echo "---------------------------------"
echo "${GET_APPS_FROM_COMPASS_RESULT}"
echo "---------------------------------"

GET_APP_COUNT=$(echo -E ${GET_APPS_FROM_COMPASS_RESULT} | jq -r '.data.result.totalCount')
compare_values 1 "${GET_APP_COUNT}" "Applications count did not match. Expected: 1. From Compass recieved: ${GET_APP_COUNT}"


GET_APP=$(echo -E ${GET_APPS_FROM_COMPASS_RESULT} | jq -c --arg appid ${CRT_APP_ID} '.data.result.data[] | select(.id==$appid)')
check_value "${GET_APP}" "Applicaiton with ID: ${CRT_APP_ID} not found in Compass"

echo "Application from Compass:"
echo "---------------------------------"
echo "${GET_APP}"
echo "---------------------------------"

GET_APP_ID=$(echo -E ${GET_APP} | jq '.id')
check_value "${GET_APP}" "Applicaiton with ID: ${CRT_APP_ID} not found in Compass"

GET_BUNDLES_COUNT=$(echo -E ${GET_APP} | jq -r '.bundles.data | length')
compare_values "${CRT_BUNDLES_COUNT}" "${GET_BUNDLES_COUNT}" "Applicaiton bundles count did not match. On creation: ${CRT_BUNDLES_COUNT}. From Compass get: ${GET_BUNDLES_COUNT}"

GET_BUNDLE_0=$(echo -E ${GET_APP} | jq -c --arg bundleid ${CRT_BUNDLE_0_ID} '.bundles.data[] | select(.id==$bundleid)')
check_value "${GET_BUNDLE_0}" "Bundle with ID: ${CRT_BUNDLE_0_ID} not found in Compass"

GET_BUNDLE_0=$(echo -E ${GET_APP} | jq -c --arg bundleid ${CRT_BUNDLE_0_ID} '.bundles.data[] | select(.id==$bundleid)')
check_value "${GET_BUNDLE_0}" "Bundle with ID: ${CRT_BUNDLE_0_ID} not found in Compass"

GET_BUNDLE_0_ID=$(echo -E ${GET_BUNDLE_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_ID}" "${GET_BUNDLE_0_ID}" "Bundle IDs did not match. On creation: ${CRT_BUNDLE_0_ID}. From Compass get: ${GET_BUNDLE_0_ID}"

GET_BUNDLE_0_API_DEFS_COUNT=$(echo -E ${GET_BUNDLE_0} | jq -r '.apiDefinitions.data | length')
compare_values "${CRT_BUNDLE_0_API_DEFS_COUNT}" "${GET_BUNDLE_0_API_DEFS_COUNT}" "Applicaiton bundles API Definitions count did not match. On creation: ${CRT_BUNDLE_0_API_DEFS_COUNT}. From Compass get: ${GET_BUNDLE_0_API_DEFS_COUNT}"

GET_BUNDLE_0_API_DEF_0=$(echo -E ${GET_BUNDLE_0} | jq -c --arg apidefid ${CRT_BUNDLE_0_API_DEF_0_ID} '.apiDefinitions.data[] | select(.id==$apidefid)')
check_value "${GET_BUNDLE_0_API_DEF_0}" "API Def with ID: ${CRT_BUNDLE_0_API_DEF_0_ID} not found in Compass"

GET_BUNDLE_0_API_DEF_0_ID=$(echo -E ${GET_BUNDLE_0_API_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_API_DEF_0_ID}" "${GET_BUNDLE_0_API_DEF_0_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_API_DEF_0_ID}. From Compass get: ${GET_BUNDLE_0_API_DEF_0_ID}"

GET_BUNDLE_0_API_DEF_1=$(echo -E ${GET_BUNDLE_0} | jq -c --arg apidefid ${CRT_BUNDLE_0_API_DEF_1_ID} '.apiDefinitions.data[] | select(.id==$apidefid)')
check_value "${GET_BUNDLE_0_API_DEF_1}" "API Def with ID: ${CRT_BUNDLE_0_API_DEF_1_ID} not found in Compass"

GET_BUNDLE_0_API_DEF_1_ID=$(echo -E ${GET_BUNDLE_0_API_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_API_DEF_1_ID}" "${GET_BUNDLE_0_API_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_API_DEF_1_ID}. From Compass get: ${GET_BUNDLE_0_API_DEF_1_ID}"

GET_BUNDLE_0_API_DEF_2=$(echo -E ${GET_BUNDLE_0} | jq -c --arg apidefid ${CRT_BUNDLE_0_API_DEF_2_ID} '.apiDefinitions.data[] | select(.id==$apidefid)')
check_value "${GET_BUNDLE_0_API_DEF_2}" "API Def with ID: ${CRT_BUNDLE_0_API_DEF_2_ID} not found in Compass"

GET_BUNDLE_0_API_DEF_2_ID=$(echo -E ${GET_BUNDLE_0_API_DEF_2} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_API_DEF_2_ID}" "${GET_BUNDLE_0_API_DEF_2_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_API_DEF_2_ID}. From Compass get: ${GET_BUNDLE_0_API_DEF_2_ID}"

GET_BUNDLE_0_EVENT_DEFS_COUNT=$(echo -E ${GET_BUNDLE_0} | jq -r '.eventDefinitions.data | length')
compare_values "${CRT_BUNDLE_0_EVENT_DEFS_COUNT}" "${GET_BUNDLE_0_EVENT_DEFS_COUNT}" "Applicaiton bundles Events Definitions count did not match. On creation: ${CRT_BUNDLE_0_EVENT_DEFS_COUNT}. From Compass get: ${GET_BUNDLE_0_EVENT_DEFS_COUNT}"

GET_BUNDLE_0_EVENT_DEF_0=$(echo -E ${GET_BUNDLE_0} | jq -c --arg eventdefid ${CRT_BUNDLE_0_EVENT_DEF_0_ID} '.eventDefinitions.data[] | select(.id==$eventdefid)')
check_value "${GET_BUNDLE_0_EVENT_DEF_0}" "API Def with ID: ${CRT_BUNDLE_0_EVENT_DEF_0_ID} not found in Compass"

GET_BUNDLE_0_EVENT_DEF_0_ID=$(echo -E ${GET_BUNDLE_0_EVENT_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_EVENT_DEF_0_ID}" "${GET_BUNDLE_0_EVENT_DEF_0_ID}" "Applicaiton bundles Event Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_EVENT_DEF_0_ID}. From Compass get: ${GET_BUNDLE_0_EVENT_DEF_0_ID}"

GET_BUNDLE_0_EVENT_DEF_1=$(echo -E ${GET_BUNDLE_0} | jq -c --arg eventdefid ${CRT_BUNDLE_0_EVENT_DEF_1_ID} '.eventDefinitions.data[] | select(.id==$eventdefid)')
check_value "${GET_BUNDLE_0_EVENT_DEF_1}" "API Def with ID: ${CRT_BUNDLE_0_EVENT_DEF_1_ID} not found in Compass"

GET_BUNDLE_0_EVENT_DEF_1_ID=$(echo -E ${GET_BUNDLE_0_EVENT_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_EVENT_DEF_1_ID}" "${GET_BUNDLE_0_EVENT_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_EVENT_DEF_1_ID}. From Compass get: ${GET_BUNDLE_0_EVENT_DEF_1_ID}"

GET_BUNDLE_1=$(echo -E ${GET_APP} | jq -c --arg bundleid ${CRT_BUNDLE_1_ID} '.bundles.data[] | select(.id==$bundleid)')
check_value "${GET_BUNDLE_1}" "Bundle with ID: ${CRT_BUNDLE_1_ID} not found in Compass"

GET_BUNDLE_1=$(echo -E ${GET_APP} | jq -c --arg bundleid ${CRT_BUNDLE_1_ID} '.bundles.data[] | select(.id==$bundleid)')
check_value "${GET_BUNDLE_1}" "Bundle with ID: ${CRT_BUNDLE_1_ID} not found in Compass"

GET_BUNDLE_1_ID=$(echo -E ${GET_BUNDLE_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_ID}" "${GET_BUNDLE_1_ID}" "Bundle IDs did not match. On creation: ${CRT_BUNDLE_1_ID}. From Compass get: ${GET_BUNDLE_1_ID}"

GET_BUNDLE_1_API_DEFS_COUNT=$(echo -E ${GET_BUNDLE_1} | jq -r '.apiDefinitions.data | length')
compare_values "${CRT_BUNDLE_1_API_DEFS_COUNT}" "${GET_BUNDLE_1_API_DEFS_COUNT}" "Applicaiton bundles API Definitions count did not match. On creation: ${CRT_BUNDLE_1_API_DEFS_COUNT}. From Compass get: ${GET_BUNDLE_1_API_DEFS_COUNT}"

GET_BUNDLE_1_API_DEF_0=$(echo -E ${GET_BUNDLE_1} | jq -c --arg apidefid ${CRT_BUNDLE_1_API_DEF_0_ID} '.apiDefinitions.data[] | select(.id==$apidefid)')
check_value "${GET_BUNDLE_1_API_DEF_0}" "API Def with ID: ${CRT_BUNDLE_1_API_DEF_0_ID} not found in Compass"

GET_BUNDLE_1_API_DEF_0_ID=$(echo -E ${GET_BUNDLE_1_API_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_API_DEF_0_ID}" "${GET_BUNDLE_1_API_DEF_0_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_API_DEF_0_ID}. From Compass get: ${GET_BUNDLE_1_API_DEF_0_ID}"

GET_BUNDLE_1_API_DEF_1=$(echo -E ${GET_BUNDLE_1} | jq -c --arg apidefid ${CRT_BUNDLE_1_API_DEF_1_ID} '.apiDefinitions.data[] | select(.id==$apidefid)')
check_value "${GET_BUNDLE_1_API_DEF_1}" "API Def with ID: ${CRT_BUNDLE_1_API_DEF_1_ID} not found in Compass"

GET_BUNDLE_1_API_DEF_1_ID=$(echo -E ${GET_BUNDLE_1_API_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_API_DEF_1_ID}" "${GET_BUNDLE_1_API_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_API_DEF_1_ID}. From Compass get: ${GET_BUNDLE_1_API_DEF_1_ID}"

GET_BUNDLE_1_API_DEF_2=$(echo -E ${GET_BUNDLE_1} | jq -c --arg apidefid ${CRT_BUNDLE_1_API_DEF_2_ID} '.apiDefinitions.data[] | select(.id==$apidefid)')
check_value "${GET_BUNDLE_1_API_DEF_2}" "API Def with ID: ${CRT_BUNDLE_1_API_DEF_2_ID} not found in Compass"

GET_BUNDLE_1_API_DEF_2_ID=$(echo -E ${GET_BUNDLE_1_API_DEF_2} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_API_DEF_2_ID}" "${GET_BUNDLE_1_API_DEF_2_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_API_DEF_2_ID}. From Compass get: ${GET_BUNDLE_1_API_DEF_2_ID}"

GET_BUNDLE_1_EVENT_DEFS_COUNT=$(echo -E ${GET_BUNDLE_1} | jq -r '.eventDefinitions.data | length')
compare_values "${CRT_BUNDLE_1_EVENT_DEFS_COUNT}" "${GET_BUNDLE_1_EVENT_DEFS_COUNT}" "Applicaiton bundles Events Definitions count did not match. On creation: ${CRT_BUNDLE_1_EVENT_DEFS_COUNT}. From Compass get: ${GET_BUNDLE_1_EVENT_DEFS_COUNT}"

GET_BUNDLE_1_EVENT_DEF_0=$(echo -E ${GET_BUNDLE_1} | jq -c --arg eventdefid ${CRT_BUNDLE_1_EVENT_DEF_0_ID} '.eventDefinitions.data[] | select(.id==$eventdefid)')
check_value "${GET_BUNDLE_1_EVENT_DEF_0}" "API Def with ID: ${CRT_BUNDLE_1_EVENT_DEF_0_ID} not found in Compass"

GET_BUNDLE_1_EVENT_DEF_0_ID=$(echo -E ${GET_BUNDLE_1_EVENT_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_EVENT_DEF_0_ID}" "${GET_BUNDLE_1_EVENT_DEF_0_ID}" "Applicaiton bundles Event Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_EVENT_DEF_0_ID}. From Compass get: ${GET_BUNDLE_1_EVENT_DEF_0_ID}"

GET_BUNDLE_1_EVENT_DEF_1=$(echo -E ${GET_BUNDLE_1} | jq -c --arg eventdefid ${CRT_BUNDLE_1_EVENT_DEF_1_ID} '.eventDefinitions.data[] | select(.id==$eventdefid)')
check_value "${GET_BUNDLE_1_EVENT_DEF_1}" "API Def with ID: ${CRT_BUNDLE_1_EVENT_DEF_1_ID} not found in Compass"

GET_BUNDLE_1_EVENT_DEF_1_ID=$(echo -E ${GET_BUNDLE_1_EVENT_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_EVENT_DEF_1_ID}" "${GET_BUNDLE_1_EVENT_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_EVENT_DEF_1_ID}. From Compass get: ${GET_BUNDLE_1_EVENT_DEF_1_ID}"


GET_BUNDLES_FROM_ORD_RESULT=$(curl --request GET --url "${ORD_URL}/open-resource-discovery-service/v0/systemInstances?%24expand=consumptionBundles(%24expand%3Dapis%2Cevents)&%24format=json" --header "authorization: Bearer ${DIRECTOR_TOKEN}" --header "tenant: ${INTERNAL_TENANT_ID}")

echo "Result from get bundles request:"
echo "---------------------------------"
echo "${GET_BUNDLES_FROM_ORD_RESULT}"
echo "---------------------------------"

ORD_APP_COUNT=$(echo -E ${GET_BUNDLES_FROM_ORD_RESULT} | jq -r '.value | length')
ORD_APP=$(echo -E ${GET_BUNDLES_FROM_ORD_RESULT} | jq -c --arg appid ${CRT_APP_ID} '.value[] | select(.id==$appid)')
check_value "${ORD_APP}" "Applicaiton with ID: ${CRT_APP_ID} not found in ORD service"

echo "Application from ORD:"
echo "---------------------------------"
echo "${ORD_APP}"
echo "---------------------------------"

ORD_BUNDLES_COUNT=$(echo -E ${ORD_APP} | jq -r '.consumptionBundles | length')
compare_values "${CRT_BUNDLES_COUNT}" "${ORD_BUNDLES_COUNT}" "Applicaiton bundles count did not match. On creation: ${CRT_BUNDLES_COUNT}. From ORD service get: ${ORD_BUNDLES_COUNT}"

ORD_BUNDLE_0=$(echo -E ${ORD_APP} | jq -c --arg bundleid ${CRT_BUNDLE_0_ID} '.consumptionBundles[] | select(.id==$bundleid)')
check_value "${ORD_BUNDLE_0}" "Bundle with ID: ${CRT_BUNDLE_0_ID} not found in ORD service"

ORD_BUNDLE_0=$(echo -E ${ORD_APP} | jq -c --arg bundleid ${CRT_BUNDLE_0_ID} '.consumptionBundles[] | select(.id==$bundleid)')
check_value "${ORD_BUNDLE_0}" "Bundle with ID: ${CRT_BUNDLE_0_ID} not found in ORD service"

ORD_BUNDLE_0_ID=$(echo -E ${ORD_BUNDLE_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_ID}" "${ORD_BUNDLE_0_ID}" "Bundle IDs did not match. On creation: ${CRT_BUNDLE_0_ID}. From ORD service get: ${ORD_BUNDLE_0_ID}"

ORD_BUNDLE_0_API_DEFS_COUNT=$(echo -E ${ORD_BUNDLE_0} | jq -r '.apis | length')
compare_values "${CRT_BUNDLE_0_API_DEFS_COUNT}" "${ORD_BUNDLE_0_API_DEFS_COUNT}" "Applicaiton bundles API Definitions count did not match. On creation: ${CRT_BUNDLE_0_API_DEFS_COUNT}. From ORD service get: ${ORD_BUNDLE_0_API_DEFS_COUNT}"

ORD_BUNDLE_0_API_DEF_0=$(echo -E ${ORD_BUNDLE_0} | jq -c --arg apidefid ${CRT_BUNDLE_0_API_DEF_0_ID} '.apis[] | select(.id==$apidefid)')
check_value "${ORD_BUNDLE_0_API_DEF_0}" "API Def with ID: ${CRT_BUNDLE_0_API_DEF_0_ID} not found in ORD service"

ORD_BUNDLE_0_API_DEF_0_ID=$(echo -E ${ORD_BUNDLE_0_API_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_API_DEF_0_ID}" "${ORD_BUNDLE_0_API_DEF_0_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_API_DEF_0_ID}. From ORD service get: ${ORD_BUNDLE_0_API_DEF_0_ID}"

ORD_BUNDLE_0_API_DEF_1=$(echo -E ${ORD_BUNDLE_0} | jq -c --arg apidefid ${CRT_BUNDLE_0_API_DEF_1_ID} '.apis[] | select(.id==$apidefid)')
check_value "${ORD_BUNDLE_0_API_DEF_1}" "API Def with ID: ${CRT_BUNDLE_0_API_DEF_1_ID} not found in ORD service"

ORD_BUNDLE_0_API_DEF_1_ID=$(echo -E ${ORD_BUNDLE_0_API_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_API_DEF_1_ID}" "${ORD_BUNDLE_0_API_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_API_DEF_1_ID}. From ORD service get: ${ORD_BUNDLE_0_API_DEF_1_ID}"

ORD_BUNDLE_0_API_DEF_2=$(echo -E ${ORD_BUNDLE_0} | jq -c --arg apidefid ${CRT_BUNDLE_0_API_DEF_2_ID} '.apis[] | select(.id==$apidefid)')
check_value "${ORD_BUNDLE_0_API_DEF_2}" "API Def with ID: ${CRT_BUNDLE_0_API_DEF_2_ID} not found in ORD service"

ORD_BUNDLE_0_API_DEF_2_ID=$(echo -E ${ORD_BUNDLE_0_API_DEF_2} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_API_DEF_2_ID}" "${ORD_BUNDLE_0_API_DEF_2_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_API_DEF_2_ID}. From ORD service get: ${ORD_BUNDLE_0_API_DEF_2_ID}"

ORD_BUNDLE_0_EVENT_DEFS_COUNT=$(echo -E ${ORD_BUNDLE_0} | jq -r '.events | length')
compare_values "${CRT_BUNDLE_0_EVENT_DEFS_COUNT}" "${ORD_BUNDLE_0_EVENT_DEFS_COUNT}" "Applicaiton bundles Events Definitions count did not match. On creation: ${CRT_BUNDLE_0_EVENT_DEFS_COUNT}. From ORD service get: ${ORD_BUNDLE_0_EVENT_DEFS_COUNT}"

ORD_BUNDLE_0_EVENT_DEF_0=$(echo -E ${ORD_BUNDLE_0} | jq -c --arg eventdefid ${CRT_BUNDLE_0_EVENT_DEF_0_ID} '.events[] | select(.id==$eventdefid)')
check_value "${ORD_BUNDLE_0_EVENT_DEF_0}" "API Def with ID: ${CRT_BUNDLE_0_EVENT_DEF_0_ID} not found in ORD service"

ORD_BUNDLE_0_EVENT_DEF_0_ID=$(echo -E ${ORD_BUNDLE_0_EVENT_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_EVENT_DEF_0_ID}" "${ORD_BUNDLE_0_EVENT_DEF_0_ID}" "Applicaiton bundles Event Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_EVENT_DEF_0_ID}. From ORD service get: ${ORD_BUNDLE_0_EVENT_DEF_0_ID}"

ORD_BUNDLE_0_EVENT_DEF_1=$(echo -E ${ORD_BUNDLE_0} | jq -c --arg eventdefid ${CRT_BUNDLE_0_EVENT_DEF_1_ID} '.events[] | select(.id==$eventdefid)')
check_value "${ORD_BUNDLE_0_EVENT_DEF_1}" "API Def with ID: ${CRT_BUNDLE_0_EVENT_DEF_1_ID} not found in ORD service"

ORD_BUNDLE_0_EVENT_DEF_1_ID=$(echo -E ${ORD_BUNDLE_0_EVENT_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_0_EVENT_DEF_1_ID}" "${ORD_BUNDLE_0_EVENT_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_0_EVENT_DEF_1_ID}. From ORD service get: ${ORD_BUNDLE_0_EVENT_DEF_1_ID}"

ORD_BUNDLE_1=$(echo -E ${ORD_APP} | jq -c --arg bundleid ${CRT_BUNDLE_1_ID} '.consumptionBundles[] | select(.id==$bundleid)')
check_value "${ORD_BUNDLE_1}" "Bundle with ID: ${CRT_BUNDLE_1_ID} not found in ORD service"

ORD_BUNDLE_1=$(echo -E ${ORD_APP} | jq -c --arg bundleid ${CRT_BUNDLE_1_ID} '.consumptionBundles[] | select(.id==$bundleid)')
check_value "${ORD_BUNDLE_1}" "Bundle with ID: ${CRT_BUNDLE_1_ID} not found in ORD service"

ORD_BUNDLE_1_ID=$(echo -E ${ORD_BUNDLE_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_ID}" "${ORD_BUNDLE_1_ID}" "Bundle IDs did not match. On creation: ${CRT_BUNDLE_1_ID}. From ORD service get: ${ORD_BUNDLE_1_ID}"

ORD_BUNDLE_1_API_DEFS_COUNT=$(echo -E ${ORD_BUNDLE_1} | jq '.apis | length')
compare_values "${CRT_BUNDLE_1_API_DEFS_COUNT}" "${ORD_BUNDLE_1_API_DEFS_COUNT}" "Applicaiton bundles API Definitions count did not match. On creation: ${CRT_BUNDLE_1_API_DEFS_COUNT}. From ORD service get: ${ORD_BUNDLE_1_API_DEFS_COUNT}"

ORD_BUNDLE_1_API_DEF_0=$(echo -E ${ORD_BUNDLE_1} | jq -c --arg apidefid ${CRT_BUNDLE_1_API_DEF_0_ID} '.apis[] | select(.id==$apidefid)')
check_value "${ORD_BUNDLE_1_API_DEF_0}" "API Def with ID: ${CRT_BUNDLE_1_API_DEF_0_ID} not found in ORD service"

ORD_BUNDLE_1_API_DEF_0_ID=$(echo -E ${ORD_BUNDLE_1_API_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_API_DEF_0_ID}" "${ORD_BUNDLE_1_API_DEF_0_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_API_DEF_0_ID}. From ORD service get: ${ORD_BUNDLE_1_API_DEF_0_ID}"

ORD_BUNDLE_1_API_DEF_1=$(echo -E ${ORD_BUNDLE_1} | jq -c --arg apidefid ${CRT_BUNDLE_1_API_DEF_1_ID} '.apis[] | select(.id==$apidefid)')
check_value "${ORD_BUNDLE_1_API_DEF_1}" "API Def with ID: ${CRT_BUNDLE_1_API_DEF_1_ID} not found in ORD service"

ORD_BUNDLE_1_API_DEF_1_ID=$(echo -E ${ORD_BUNDLE_1_API_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_API_DEF_1_ID}" "${ORD_BUNDLE_1_API_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_API_DEF_1_ID}. From ORD service get: ${ORD_BUNDLE_1_API_DEF_1_ID}"

ORD_BUNDLE_1_API_DEF_2=$(echo -E ${ORD_BUNDLE_1} | jq -c --arg apidefid ${CRT_BUNDLE_1_API_DEF_2_ID} '.apis[] | select(.id==$apidefid)')
check_value "${ORD_BUNDLE_1_API_DEF_2}" "API Def with ID: ${CRT_BUNDLE_1_API_DEF_2_ID} not found in ORD service"

ORD_BUNDLE_1_API_DEF_2_ID=$(echo -E ${ORD_BUNDLE_1_API_DEF_2} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_API_DEF_2_ID}" "${ORD_BUNDLE_1_API_DEF_2_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_API_DEF_2_ID}. From ORD service get: ${ORD_BUNDLE_1_API_DEF_2_ID}"

ORD_BUNDLE_1_EVENT_DEFS_COUNT=$(echo -E ${ORD_BUNDLE_1} | jq -r '.events | length')
compare_values "${CRT_BUNDLE_1_EVENT_DEFS_COUNT}" "${ORD_BUNDLE_1_EVENT_DEFS_COUNT}" "Applicaiton bundles Events Definitions count did not match. On creation: ${CRT_BUNDLE_1_EVENT_DEFS_COUNT}. From ORD service get: ${ORD_BUNDLE_1_EVENT_DEFS_COUNT}"

ORD_BUNDLE_1_EVENT_DEF_0=$(echo -E ${ORD_BUNDLE_1} | jq -c --arg eventdefid ${CRT_BUNDLE_1_EVENT_DEF_0_ID} '.events[] | select(.id==$eventdefid)')
check_value "${ORD_BUNDLE_1_EVENT_DEF_0}" "API Def with ID: ${CRT_BUNDLE_1_EVENT_DEF_0_ID} not found in ORD service"

ORD_BUNDLE_1_EVENT_DEF_0_ID=$(echo -E ${ORD_BUNDLE_1_EVENT_DEF_0} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_EVENT_DEF_0_ID}" "${ORD_BUNDLE_1_EVENT_DEF_0_ID}" "Applicaiton bundles Event Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_EVENT_DEF_0_ID}. From ORD service get: ${ORD_BUNDLE_1_EVENT_DEF_0_ID}"

ORD_BUNDLE_1_EVENT_DEF_1=$(echo -E ${ORD_BUNDLE_1} | jq -c --arg eventdefid ${CRT_BUNDLE_1_EVENT_DEF_1_ID} '.events[] | select(.id==$eventdefid)')
check_value "${ORD_BUNDLE_1_EVENT_DEF_1}" "API Def with ID: ${CRT_BUNDLE_1_EVENT_DEF_1_ID} not found in ORD service"

ORD_BUNDLE_1_EVENT_DEF_1_ID=$(echo -E ${ORD_BUNDLE_1_EVENT_DEF_1} | jq -r '.id')
compare_values "${CRT_BUNDLE_1_EVENT_DEF_1_ID}" "${ORD_BUNDLE_1_EVENT_DEF_1_ID}" "Applicaiton bundles API Definitions IDs did not match. On creation: ${CRT_BUNDLE_1_EVENT_DEF_1_ID}. From ORD service get: ${ORD_BUNDLE_1_EVENT_DEF_1_ID}"

echo "Ord-test end reached. Test finished with ${TEST_RESULT}!"

echo "Wait 5s before collect logs ..."
sleep 5

echo "Logs from Director:"
echo "---------------------------------"
cat "${COMPASS_DIR}/components/director/main.log" || true
echo "---------------------------------"

echo "Logs from ORD Service:"
echo "---------------------------------"
cat "${ORD_SVC_DIR}/components/ord-service/target/main.log" || true
echo "---------------------------------"

if [[ ${TEST_RESULT} == false ]]; then
    echo "Test Fail. Look for COMPARE ERROR or VALIDATION ERROR messages."
    exit 1
else
    echo "Test Pass"
fi

