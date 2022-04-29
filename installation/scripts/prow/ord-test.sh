#!/bin/bash

###
# Following script installs necessary tooling for Debian, starts Director and Ord service, and runs the smoke tests.
#

set -o errexit

function is_ready(){
    local URL=${1}
    local HTTP_CODE=$(curl -s -o /dev/null -I -w "%{http_code}" ${URL})
    if [[ "${HTTP_CODE}" == "200" ]]; then
        return 0
    fi 
    echo "Response from ${URL} is still: ${HTTP_CODE}"
    return 1
}

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR="$( cd "$( dirname "${CURRENT_DIR}/../../.." )" && pwd )"
BASE_DIR="$( cd "$( dirname "${INSTALLATION_DIR}/../../../../../../.." )" && pwd )"
JAVA_HOME="${BASE_DIR}/openjdk-11"
M2_HOME="${BASE_DIR}/maven"
MIGRATE_HOME="${BASE_DIR}/migrate"
COMPASS_DIR="$( cd "$( dirname "${INSTALLATION_DIR}/../.." )" && pwd )"  
ORD_SVC_DIR="${BASE_DIR}/ord-service"

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
cd ${COMPASS_DIR}/components/director
./run.sh &

echo "Wait compass to start for 60 seconds ..."
sleep 60

COMPASS_URL="http://localhost:3000"

STARTE_TIME=$(date +%s)
until is_ready "${COMPASS_URL}/healthz" ; do
    CURRENT_TME=$(date +%s)
    SECONDS=$((CURRENT_TME-STARTE_TIME))
    if (( SECONDS > 300 )); then
        echo "Timeout of 5 min for starting compass reached. Exiting."
        exit 1
    fi
    echo "Wait 5s ..."
    sleep 5
done

. ${COMPASS_DIR}/components/director/hack/jwt_generator.sh

read -r DIRECTOR_TOKEN <<< $(get_token)
read -r INTERNAL_TENANT_ID <<< $(get_internal_tenant)
echo "Compass is ready"

echo "Starting ord-service"
cd ${ORD_SVC_DIR}/components/ord-service
export SERVER_PORT=8081
./run.sh --migrations-path ${COMPASS_DIR}/components/schema-migrator/migrations/director &

echo "Wait ord-service to start for 60 seconds ..."
sleep 60

ORD_URL="http://localhost:${SERVER_PORT}"

STARTE_TIME=$(date +%s)
until is_ready "${ORD_URL}/actuator/health" ; do
    CURRENT_TME=$(date +%s)
    SECONDS=$((CURRENT_TME-STARTE_TIME))
    if (( SECONDS > 300 )); then
        echo "Timeout of 5 min for starting ord-service reached. Exiting."
        exit 1
    fi
    echo "Wait 5s ..."
    sleep 5
done

echo "ORD-service is ready"

echo "Token: ${DIRECTOR_TOKEN}"
echo "Internal Tenant ID: ${INTERNAL_TENANT_ID}"

echo "ord-test end reached!"
