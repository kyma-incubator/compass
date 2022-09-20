#!/usr/bin/env bash

# This script is responsible for running Director with PostgreSQL.

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -e

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

SKIP_DB_CLEANUP=false
REUSE_DB=false
DUMP_DB=false
AUTO_TERMINATE=false
DISABLE_ASYNC_MODE=true
COMPONENT='director'
TERMINAION_TIMEOUT_IN_SECONDS=300

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
            DEBUG_PORT=40000
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
          export APP_SYSTEM_TO_TEMPLATE_MAPPINGS='[{  "Name": "SAP S/4HANA On-Premise",  "SourceKey": ["type"],  "SourceValue": ["on-premise"]}]'
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
        --auto-terminate)
             AUTO_TERMINATE=true
             TERMINAION_TIMEOUT_IN_SECONDS=$2
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

POSTGRES_CONTAINER="test-postgres"
POSTGRES_VERSION="11"

DB_USER="postgres"
DB_PWD="pgsql@12345"
DB_NAME="compass"
DB_PORT="5432"
DB_HOST="127.0.0.1"

CLIENT_CERT_SECRET_NAMESPACE="default"
CLIENT_CERT_SECRET_NAME="external-client-certificate"

function cleanup() {

    if [[ ${DEBUG} == true ]]; then
       echo -e "${GREEN}Cleanup Director binary${NC}"
       rm  $GOPATH/src/github.com/kyma-incubator/compass/components/director/director
    fi

    if [[ ${SKIP_DB_CLEANUP} = false ]]; then
        echo -e "${GREEN}Cleanup Postgres container${NC}"
        docker rm --force ${POSTGRES_CONTAINER}
    else
        echo -e "${GREEN}Skipping Postgres container cleanup${NC}"
    fi

    echo -e "${GREEN}Destroying k3d cluster...${NC}"
    k3d cluster delete k3d-cluster
}

trap cleanup EXIT

echo -e "${GREEN}Creating k3d cluster...${NC}"
curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v5.2.0 bash
k3d cluster create k3d-cluster --api-port 6550 --servers 1 --port 443:443@loadbalancer --image rancher/k3s:v1.22.4-k3s1 --kubeconfig-update-default --wait

if [[ ${REUSE_DB} = true ]]; then
    echo -e "${GREEN}Will reuse existing Postgres container${NC}"
else
    set +e
    echo -e "${GREEN}Start Postgres in detached mode${NC}"
    docker run -d --name ${POSTGRES_CONTAINER} \
                -e POSTGRES_HOST=${DB_HOST} \
                -e POSTGRES_USER=${DB_USER} \
                -e POSTGRES_PASSWORD=${DB_PWD} \
                -e POSTGRES_DB=${DB_NAME} \
                -e POSTGRES_PORT=${DB_PORT} \
                -p ${DB_PORT}:${DB_PORT} \
                postgres:${POSTGRES_VERSION}

    if [[ $? -ne 0 ]] ; then
        SKIP_DB_CLEANUP=true
        exit 1
    fi

    echo '# WAITING FOR CONNECTION WITH DATABASE #'
    for i in {1..30}
    do
        docker exec ${POSTGRES_CONTAINER} pg_isready -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"
        if [ $? -eq 0 ]
        then
            dbReady=true
            break
        fi
        sleep 1
    done

    if [ "${dbReady}" != true ] ; then
        echo '# COULD NOT ESTABLISH CONNECTION TO DATABASE #'
        exit 1
    fi

    set -e

    echo -e "${GREEN}Populate DB${NC}"

    if [[ ${DUMP_DB} = false ]]; then
        CONNECTION_STRING="postgres://$DB_USER:$DB_PWD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
        migrate -path ${ROOT_PATH}/../schema-migrator/migrations/director -database "$CONNECTION_STRING" up

        cat ${ROOT_PATH}/../schema-migrator/seeds/director/*.sql | \
            docker exec -i ${POSTGRES_CONTAINER} psql -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"
    else
        if [[ ! -f ${ROOT_PATH}/../schema-migrator/seeds/dump.sql ]]; then
            echo -e "${GREEN}Will pull DB dump from GCR bucket${NC}"
            gsutil cp gs://sap-cp-cmp-dev-db-dump/dump.sql ${ROOT_PATH}/../schema-migrator/seeds/dump.sql
        fi

        cat ${ROOT_PATH}/../schema-migrator/seeds/dump.sql | \
            docker exec -i ${POSTGRES_CONTAINER} psql -v ON_ERROR_STOP=1 -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}"

        REMOTE_MIGRATION_VERSION=$(docker exec -i ${POSTGRES_CONTAINER} psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT version FROM schema_migrations")
        LOCAL_MIGRATION_VERSION=$(echo $(ls ${ROOT_PATH}/../schema-migrator/migrations/director | tail -n 1) | grep -o -E '[0-9]+' | head -1 | sed -e 's/^0\+//')

        if [[ ${REMOTE_MIGRATION_VERSION} = ${LOCAL_MIGRATION_VERSION} ]]; then
            echo -e "${GREEN}Both remote and local migrations are at the same version.${NC}"
        else
            echo -e "${YELLOW}NOTE: Remote and local migrations are at different versions.${NC}"
            echo -e "${YELLOW}REMOTE:${NC} $REMOTE_MIGRATION_VERSION"
            echo -e "${YELLOW}LOCAL:${NC} $LOCAL_MIGRATION_VERSION"
        fi

        CONNECTION_STRING="postgres://$DB_USER:$DB_PWD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
        migrate -path ${ROOT_PATH}/../schema-migrator/migrations/director -database "$CONNECTION_STRING" up
    fi
fi

echo "Migration version: $(migrate -path ${ROOT_PATH}/../schema-migrator/migrations/director -database "$CONNECTION_STRING" version 2>&1)"
. ${ROOT_PATH}/hack/jwt_generator.sh

if [[  ${SKIP_APP_START} ]]; then
    echo -e "${GREEN}Skipping starting application${NC}"
    while true
    do
        sleep 1
    done
fi

echo -e "${GREEN}Starting application${NC}"

export APP_DB_USER=${DB_USER}
export APP_DB_PASSWORD=${DB_PWD}
export APP_DB_NAME=${DB_NAME}
export APP_CONFIGURATION_FILE=${ROOT_PATH}/hack/config-local.yaml
export APP_OAUTH20_URL="https://oauth2-admin.kyma.local"
export APP_OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT="https://oauth2.kyma.local/oauth2/token"
export APP_ONE_TIME_TOKEN_URL="http://connector.not.configured.url/graphql"
export APP_URL="http://director.not.configured.url/director"
export APP_CONNECTOR_URL="http://connector.not.configured.url/connector/graphql"
export APP_LEGACY_CONNECTOR_URL="https://adapter-gateway.kyma.local/v1/applications/signingRequests/info"
export APP_LOG_LEVEL=debug
export APP_HTTP_RETRY_ATTEMPTS=3
export APP_HTTP_RETRY_DELAY=100ms
export APP_DISABLE_ASYNC_MODE=${DISABLE_ASYNC_MODE}
export APP_DISABLE_TENANT_ON_DEMAND_MODE=true
export APP_HEALTH_CONFIG_INDICATORS="{database,5s,1s,1s,3}"
export APP_SUGGEST_TOKEN_HTTP_HEADER=suggest_token
export APP_SCHEMA_MIGRATION_VERSION=$(ls -lr ${ROOT_PATH}/../schema-migrator/migrations/director | head -n 2 | tail -n 1 | tr -s ' ' | cut -d ' ' -f9 | cut -d '_' -f1)
export APP_ALLOW_JWT_SIGNING_NONE=true
export APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY="non-existent-label-key"
export APP_EXTERNAL_CLIENT_CERT_SECRET=${CLIENT_CERT_SECRET_NAMESPACE}/${CLIENT_CERT_SECRET_NAME}
export APP_EXTERNAL_CLIENT_CERT_KEY="tls.crt"
export APP_EXTERNAL_CLIENT_KEY_KEY="tls.key"
export APP_EXTERNAL_CLIENT_CERT_VALUE="certValue" # the default value is not valid but if you want you can override with the desired certificate value
export APP_EXTERNAL_CLIENT_KEY_VALUE="keyValue" # the default value is not valid but if you want you can override with the desired key value
export APP_INFO_ROOT_CA="--- Feature Disabled Locally ---"
export APP_SELF_REGISTER_OAUTH_X509_CERT="LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUYwakNDQTdxZ0F3SUJBZ0lKQUtudzEwQi9zejJUTUEwR0NTcUdTSWIzRFFFQkN3VUFNRTB4Q3pBSkJnTlYKQkFZVEFrSkhNUTR3REFZRFZRUUlEQVZzYjJOaGJERU9NQXdHQTFVRUJ3d0ZiRzlqWVd3eERqQU1CZ05WQkFvTQpCV3h2WTJGc01RNHdEQVlEVlFRRERBVnNiMk5oYkRBZ0Z3MHlNakF5TWpVeE16VTFNVEJhR0E4eU1USXlNREl3Ck1URXpOVFV4TUZvd1RURUxNQWtHQTFVRUJoTUNRa2N4RGpBTUJnTlZCQWdNQld4dlkyRnNNUTR3REFZRFZRUUgKREFWc2IyTmhiREVPTUF3R0ExVUVDZ3dGYkc5allXd3hEakFNQmdOVkJBTU1CV3h2WTJGc01JSUNJakFOQmdrcQpoa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQW9UOXdxSE9kVUVqNitGbWhnbHJFYXNQY2l3dzBwSzQvCjA3cDc2MVZXN0JjVk5zWXJ3eXQ3UFB4cW1sSTg3RzZPVGEzN0pRSk9IOXIxSnlVdlRJejBta3ZxazBNWmtLQk8KdjZTTER5Q3BwOGc0TUd6Q0tHcldOWUJIRmNyMHdLS2w1b1V0WG45dDBHOFNXOG1NRjFxWk9pOWlMZEVZeE9kbgpHcmFvWno5RTZ0TFRQU0FDdEJHOVBSdENwb0VQOVl5di9XSjA3ZzlMeG1LL3NZVW1wSjB6RWdscDliYVRETjU1CkNyaFA0TnNxS2R1L0tHd3ArOElTNzFwZ3BYS2hhZ2t5M3JYVVZ5bmNkbXBIS1g0bkE0R0h0S0xSMkE1OVByK1oKVlkrbFBXYXB0Tnh3WEZ6RkNBUVJSbTVib1FIRUhTUmhNbXByL3phemdaUzBobm0zSS9taUZOZjg5L3NFSWpaZwpBYjV4WnpUU0MzNUs4WkV3a3dEOS9hdUVGRHdwby9EYktySzNtSTc4cFlqd2xNKzB6d09rZE5sVHVaUVY2VmFTCkJSekw0L2lBd0tnQ1dJTTNkb1RGczZiVHcrVXlWK2xuLzdkcDFBZWxrbm5TNVpkMFhtUkh4NXVlcmhRQUZDY1kKdzlKT25BWTk1by9RUGp4RWhWaS9tS2R2Z1A2VVg5T1orVHZ1UFB6TnlmK05KeDlFTk90UDVMVk5wUHVSZzJrTQp4NlZ4bnZUeUo4LzBSNTFGOW5qZ3ZxdzF0NDRyZ0J5L3E0VWlrK1h2QUw0dVY0aWcycFUrcEVieW1qekoyZTNpCjNBSkpXQ05hZUdGa3c2YUFpTlhOZmk4VGk1RjdteDgwS2J5S0FmVlMwNDJqRUhtMHYvY3pNN2lxNXdRWkVFSUYKUDZxZDk4M0daazhDQXdFQUFhT0JzakNCcnpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJUbwprMzBXVzhCRUxnMEZBWEZscjNnYnpkektvVEI5QmdOVkhTTUVkakIwZ0JUb2szMFdXOEJFTGcwRkFYRmxyM2diCnpkektvYUZScEU4d1RURUxNQWtHQTFVRUJoTUNRa2N4RGpBTUJnTlZCQWdNQld4dlkyRnNNUTR3REFZRFZRUUgKREFWc2IyTmhiREVPTUF3R0ExVUVDZ3dGYkc5allXd3hEakFNQmdOVkJBTU1CV3h2WTJGc2dna0FxZkRYUUgregpQWk13RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUpkbHNtR2k0d1hvSlZ4SnlKVzlDWEZPejhZWkhTWlhicEdsClJXcmI4QkZIMy9SNFhPTTQ5Y3Y4UzErYUZvQ3hJd0taRGFhZVIwNVIxK05jVXNPSnhZL2tGWXNHN3kvMTJFRVIKTi90anVRTGhPNnEzT1piZTUwUFErS2pxbmxURnQrT1ptV1ZQc3NzbUU2WWNhWXBOc09FcmVWZWxNWHNybDBndgpLZ3hLUXFJQ3hJTThNTG1QeUUxSURsS3RSL0RrTkNqNDQybmgvNVlwaUFoOG1BTUE5QjFtcE1uSTZpZFB0RzRhCm5uSGxiVlZaMUE4UWF2bXlCNGRueVZwNlB3QnhSMXJjN0xoV3VBT3V2WkNWWGpxUVZLMkREdEJ0Q3RGcUUrQlQKa1I4MnFjMGtKR2IwYTFkRWRubmJEdU1BQzA0WnRrQnpQTlV6RzNqSkNKVkZhVEsyMVhZQjNkMUR4UTVSUFJyTQptZmVXUlM5cytmdlplOTZmb1BIMmZuREN0aFJkOXI5UXdWSmhZMkRlTzlQOVd0TW9QMWhMempHSnQ1czB2RmpRClE2RDdZaXFHajkvTXhoc1BqZjlucjhCL3ZUVzFWV1hackh2NDBlR01tM1kwRXUvMlA0bFR5YnEyUlp2enhlQTgKTmlXL0lEa3VrbmlMaU95UnZKL1RWVXFsYllhbm9qeUVlbzJaTnp0RTVadTRYUndabnZIeVIzNFVrMldTYjdOSAp4Tk53TUt6ZHdROFpQQnM1OCtZUjByUGlGN2V4WXprVnFOdHNydEYwUjVPRm5kZjN4Yi9GeC8rRTBhd1UyUlZSClNiTzUxM1JIMTFZZ3RneWlMM0kxRjk3NGFDZTRhYkJuWHdmWDA1eUl2a2dEbFAvdXA5T25WRjV6eU8xSkRmUDAKSlFqY20zbGsKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
export APP_SELF_REGISTER_OAUTH_X509_KEY="LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBb1Q5d3FIT2RVRWo2K0ZtaGdsckVhc1BjaXd3MHBLNC8wN3A3NjFWVzdCY1ZOc1lyCnd5dDdQUHhxbWxJODdHNk9UYTM3SlFKT0g5cjFKeVV2VEl6MG1rdnFrME1aa0tCT3Y2U0xEeUNwcDhnNE1HekMKS0dyV05ZQkhGY3Iwd0tLbDVvVXRYbjl0MEc4U1c4bU1GMXFaT2k5aUxkRVl4T2RuR3Jhb1p6OUU2dExUUFNBQwp0Qkc5UFJ0Q3BvRVA5WXl2L1dKMDdnOUx4bUsvc1lVbXBKMHpFZ2xwOWJhVERONTVDcmhQNE5zcUtkdS9LR3dwCis4SVM3MXBncFhLaGFna3kzclhVVnluY2RtcEhLWDRuQTRHSHRLTFIyQTU5UHIrWlZZK2xQV2FwdE54d1hGekYKQ0FRUlJtNWJvUUhFSFNSaE1tcHIvemF6Z1pTMGhubTNJL21pRk5mODkvc0VJalpnQWI1eFp6VFNDMzVLOFpFdwprd0Q5L2F1RUZEd3BvL0RiS3JLM21JNzhwWWp3bE0rMHp3T2tkTmxUdVpRVjZWYVNCUnpMNC9pQXdLZ0NXSU0zCmRvVEZzNmJUdytVeVYrbG4vN2RwMUFlbGtublM1WmQwWG1SSHg1dWVyaFFBRkNjWXc5Sk9uQVk5NW8vUVBqeEUKaFZpL21LZHZnUDZVWDlPWitUdnVQUHpOeWYrTkp4OUVOT3RQNUxWTnBQdVJnMmtNeDZWeG52VHlKOC8wUjUxRgo5bmpndnF3MXQ0NHJnQnkvcTRVaWsrWHZBTDR1VjRpZzJwVStwRWJ5bWp6SjJlM2kzQUpKV0NOYWVHRmt3NmFBCmlOWE5maThUaTVGN214ODBLYnlLQWZWUzA0MmpFSG0wdi9jek03aXE1d1FaRUVJRlA2cWQ5ODNHWms4Q0F3RUEKQVFLQ0FnRUFqUDBpYlRmQjZrd1ZuUWNKOENlYkxGc2JRRDBvM29FNWI5RFR2ejQ4Sld3OWNVb3ZRNVNHU2huTwp3Q1o5L0tEaUxrdWNsNHgvY04wTGsvR3dmTGVXdkQ3NjJVNUhVU3pLRGtrNkNiMGVlb1RYbElmVDhIRVI0Vy9MCk4rUGd3M3F6b203NTczRnVQRnlSNmMyOWYwSUpUbFhWKzRlanA2OUplSk1UaGt0TTRDSDg3NnBJa3RnYjVnMHEKNXRsY2NmQlVoVElNV1lib1U0dE9YMUswS2lVRlhaVDdvQXZHWWU4NFdNWTFtYjhvQzdlSFdqblJMNzlPdlJnQgovMGZPbVI5MzZrR0VhNzQvZFE2U01GYU1tRVV1dWlQUFphR3RveXIyVUZpc081YkRkazkwczEydUxjY1lyOE9ZCnZKd0Z0UkYxSnhia1hSK2dMd0l1SXBMVUxsRjhoR2ZEb3QvWUZvZGJBM3A0dlM1MEc4c212YWlROVhnVXdGdFoKMFdpWU5iQ2JKdlY1d0w2aXlPcTRqakFGeExqU2lqc20ybFpVcTZiVDY5YVFpR0xmbXJDYnJDRnJFSkNpU1F3WgpGenVrMGlXYnM4ZXRySWFnNjdtY0kvTlF4RkVQbXB2ZEdQTllGZ0RwUDdEMnlhUnA4MzVjeWJVd04vcUE2RnBGCkJHSEQ0ZWl6OWd0WlhwT25NYzBibk9KcG5qMlNpaUUwaS95a0NuTlZvRm1SeXladVY0YVlQd29laWpOZ1NjZWIKQTY1TThHaWtUdmdiemFaL3YvOVdlWElURGJVWjcvQTBwYlRDTjZZZ0xscjBpQlJPMUFmeGhNNkVmems4QlFUSQpidWJ5U3ZwQWhiRk83aTZLVVAvZW03eVV0ckxBekdOR2dnTFJMVEpRWGdYUWNSeUdVWkVDZ2dFQkFNOUVhZW5nCm9peE1QaE5jTlN0dENEekhER2RDSFRWcDN2U0s0bkNHRFhGY3N2aytqMGRRM2pQYWx2V3dDMjBYOTZWQlR1bzAKTFhhcGgwZ0xHYloyS3JWQ0tPOWZrZXpQVWMwOWMzNmp1OExFZ3NBZ0dGOUpXeDVlMUVleUIzS0w0RnIreXVVVwpLbDc0azltYmZVK1dRcG1vc3hKdW41ZjBqbzhoclRucGtqR2FqUHhTL05GaE1YVHpGdjl6dEhVblZhTzJVYzI3ClB1WUJNSkxGdTQ2Y2xxWG53RTQxUTVwVHFLZGwxaUw0dFRrVER5ekFNUHZucW9HSld4Q1ZidjR3b2Z1WXBrY0gKSG50SlVEMVhhSnI1N0xoSU4zeDRCN2pFU3JYTUwrOVVIcWgzTDRna2FwV0VTTDlsY1d3aWxlcHhYdG5RdDI3NgprK0NXVVpuY09TUklVM2NDZ2dFQkFNY3BGRGdBWHVYS040b0Z3TC9lRkkvUWVSeXBDakQ2LytidVdCV2F0SXdQCnprQ1h6ODNaY256NmQ4YUlHc2I1bi9tbFYwaHBCdkdtZzBkN0ZHbHZoWDVhZDRzeXFLZWZJajZtWG1PMDRrQ3gKa29kZkErUVh5Q1EwWlFSRi9TenRyNWxWQi9yYnhhYzdCVlplQUxvVUtSSzdUMFhCOHczVXN1SnNyNGR0QVkwZQpDUDlmQlRybFowYjBFSnROWXlVbGFSNWRkaFVuYnJFeHFkUzg5Yk8ybUxBUjVDTVJJRXZScTRseGlQYzRTTXhYCmhUY2FSTmpBU2M3WE5NUUI1WVpXeUxaYzN1RndQaXRHK0sxTGMwNWs1VFpWeEtNdmVhOXBiZkRHLysyQm9IcHIKOEFHVXVKVlk2bCtGTzNleXlmenRPVGtpMUkwNmlxU2dURzVrbzJ0bXlla0NnZ0VBUjVYZWFzdU4xM1RodjdnSwpHUnlJU3MySXFDVTZoMWN3alE5bTAreEl1azJFOXZhM2I2OHJmNGRRdWp4NlJjeVFXTUFzckZFbkhxUEF1STQwCjdFTDF6ekt4aHJOZ2FBVFd3T2NuZTZhN1U3S2hZZy96dXYxUC9qWk1aUkxFNWJnUDNmM0FQODBmQnp3ZGZIdnEKbE5GVjRWSlZ2dGo4UC9SVVJIVWlLaTFVczlNb1BJSEJGZVBXdkFpMWViY1JyYURQUUVMWkVCQkswZys1SWdndgpGanRaQUtZQlVrR3RQcUVFVUFTcEo5ejBZbWtGeGJQL2R4RjFYMVg4WU1icjFka2dLUkI0NVhFOUF1RzRWK2RYCmxxY1pMakNyRVU4M2c0WXdNNGY1U2xTb1hoRUVGcVpWTlp6QnIzRXU4bVVqbUJ4ZDRTYm9JK2xocDZEalFCdkMKbEpoeVV3S0NBUUI4QUpyREo0L3VtVkt0VUZtcjNQV0dlYkgrNDAwaUpCWFRUbEZ2Mml4U0RNRkp2SHc1V2d1TAp2MU4yUEdZWHYzTVl1QmE1VWhOdHdGUjYzQ3BnWDN5SnFJQklIaG1lakZtQkVvc3duMzVEODR3ZFYwNlA1VExMClFBZ3BlZjVoeS9nS2kwUDFzSUxIVmR0RDVER2xxa25Nak8yVnJHWE9GY0h2Y3VaemRxNkJrOUxjVmVobXZGRHEKZjZvYldEckQ5U0FYTlBBQnlkU0U1VHd0NWgxQmNRNXVpaVUycEVJc2t2YXdGQTNJaDdYajdSWlhzYlp1RW9PaQpFcUthNitkaUZvVFA3dEVqSW9UQzQyU1FXYXNJZzQrbm5nMVo0WVJ0Y0VKd3FTYk9WV2g2OE51MTBFaUJUS1JaCkp4Wll0K3hGMjlwR05lYUxySWlJYWZwTXZjSjJhOENKQW9JQkFRRE5Nbk43M1k4UWNwSld4eTk4SmNLSklrTTQKY3BIeUM1U2dva2QvRWdBZzdMdW96a0grQWtWREZTUmcxdmRDSjdSMU9yTXNxeUVyYy9xV0pDYWpqVWJ6UlVyOApJQlN5b2Y1Zko3UHZ1VFc2WkV6RUR2WEw5WW90VExLM1VnUWpGc0N6c3dIUVhOaEE1QVBHWWZZa2hsTnlubVdpCnpyYlFYU1Q2RkxDeSt3UW9UQ3NzVHF4NW8yOEZiZ21RSTZvQzdmQ0ViTzZNL2dCTUxtbjhFNW03dWJVWE9wZmYKaEhsVGcwaVpKU0lDTDNRK25aZ0J4dVNSZ1hsMjJkUjZNYndiRDR6TTltNFhjTjNLTkNGMFY2bkJuQ0dnQ2hNaQp3aWpnS0w5RU1LemZFN0N4K05ZR0x5em9icVBXWDdMak81UjlPUWlpa3FBNmxxTldlN1g0b2NacFZ6c28KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"
export APP_ORD_WEBHOOK_MAPPINGS='[{ "OrdUrlPath": "/.well-known/open-resource-discovery", "Type": "SAP temp1", "PpmsProductVersions": ["12345"], "SubdomainSuffix": "" }]'

# Tenant Fetcher properties
export APP_SUBSCRIPTION_CALLBACK_SCOPE=Callback
export APP_TENANT_REGION_DEPENDENCIES_CONFIG_PATH="/tmp/dependencies.json"

# Self Register Properties
export APP_SELF_REGISTER_SECRET_PATH="/tmp/keyConfig"
export APP_SELF_REGISTER_INSTANCE_CLIENT_ID_PATH="clientId"
export APP_SELF_REGISTER_INSTANCE_CLIENT_SECRET_PATH="clientSecret"
export APP_SELF_REGISTER_INSTANCE_URL_PATH="url"
export APP_SELF_REGISTER_INSTANCE_TOKEN_URL_PATH="tokenUrl"
export APP_SELF_REGISTER_INSTANCE_X509_CERT_PATH="clientCert"
export APP_SELF_REGISTER_INSTANCE_X509_KEY_PATH="clientKey"

# Pairing Adapters Properties
export APP_PAIRING_ADAPTER_CM_NAME="pairing-adapter-config-local"
export APP_PAIRING_ADAPTER_CM_NAMESPACE="default"
export APP_PAIRING_ADAPTER_CM_KEY="config.json"
export APP_PAIRING_ADAPTER_WATCHER_ID="pairing-adapter-watcher-id"

# This file contains necessary configuration for self registration flow
cat <<EOF > /tmp/keyConfig
{
  "eu-1": {
    "clientId": "client_id",
    "clientSecret": "client_secret",
    "url": "http://compass-external-services-mock.compass-system.svc.cluster.local:8080",
    "tokenUrl": "https://compass-external-services-mock-sap-mtls.kyma-local:8080",
    "clientCert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUYwakNDQTdxZ0F3SUJBZ0lKQUtudzEwQi9zejJUTUEwR0NTcUdTSWIzRFFFQkN3VUFNRTB4Q3pBSkJnTlYKQkFZVEFrSkhNUTR3REFZRFZRUUlEQVZzYjJOaGJERU9NQXdHQTFVRUJ3d0ZiRzlqWVd3eERqQU1CZ05WQkFvTQpCV3h2WTJGc01RNHdEQVlEVlFRRERBVnNiMk5oYkRBZ0Z3MHlNakF5TWpVeE16VTFNVEJhR0E4eU1USXlNREl3Ck1URXpOVFV4TUZvd1RURUxNQWtHQTFVRUJoTUNRa2N4RGpBTUJnTlZCQWdNQld4dlkyRnNNUTR3REFZRFZRUUgKREFWc2IyTmhiREVPTUF3R0ExVUVDZ3dGYkc5allXd3hEakFNQmdOVkJBTU1CV3h2WTJGc01JSUNJakFOQmdrcQpoa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQW9UOXdxSE9kVUVqNitGbWhnbHJFYXNQY2l3dzBwSzQvCjA3cDc2MVZXN0JjVk5zWXJ3eXQ3UFB4cW1sSTg3RzZPVGEzN0pRSk9IOXIxSnlVdlRJejBta3ZxazBNWmtLQk8KdjZTTER5Q3BwOGc0TUd6Q0tHcldOWUJIRmNyMHdLS2w1b1V0WG45dDBHOFNXOG1NRjFxWk9pOWlMZEVZeE9kbgpHcmFvWno5RTZ0TFRQU0FDdEJHOVBSdENwb0VQOVl5di9XSjA3ZzlMeG1LL3NZVW1wSjB6RWdscDliYVRETjU1CkNyaFA0TnNxS2R1L0tHd3ArOElTNzFwZ3BYS2hhZ2t5M3JYVVZ5bmNkbXBIS1g0bkE0R0h0S0xSMkE1OVByK1oKVlkrbFBXYXB0Tnh3WEZ6RkNBUVJSbTVib1FIRUhTUmhNbXByL3phemdaUzBobm0zSS9taUZOZjg5L3NFSWpaZwpBYjV4WnpUU0MzNUs4WkV3a3dEOS9hdUVGRHdwby9EYktySzNtSTc4cFlqd2xNKzB6d09rZE5sVHVaUVY2VmFTCkJSekw0L2lBd0tnQ1dJTTNkb1RGczZiVHcrVXlWK2xuLzdkcDFBZWxrbm5TNVpkMFhtUkh4NXVlcmhRQUZDY1kKdzlKT25BWTk1by9RUGp4RWhWaS9tS2R2Z1A2VVg5T1orVHZ1UFB6TnlmK05KeDlFTk90UDVMVk5wUHVSZzJrTQp4NlZ4bnZUeUo4LzBSNTFGOW5qZ3ZxdzF0NDRyZ0J5L3E0VWlrK1h2QUw0dVY0aWcycFUrcEVieW1qekoyZTNpCjNBSkpXQ05hZUdGa3c2YUFpTlhOZmk4VGk1RjdteDgwS2J5S0FmVlMwNDJqRUhtMHYvY3pNN2lxNXdRWkVFSUYKUDZxZDk4M0daazhDQXdFQUFhT0JzakNCcnpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJUbwprMzBXVzhCRUxnMEZBWEZscjNnYnpkektvVEI5QmdOVkhTTUVkakIwZ0JUb2szMFdXOEJFTGcwRkFYRmxyM2diCnpkektvYUZScEU4d1RURUxNQWtHQTFVRUJoTUNRa2N4RGpBTUJnTlZCQWdNQld4dlkyRnNNUTR3REFZRFZRUUgKREFWc2IyTmhiREVPTUF3R0ExVUVDZ3dGYkc5allXd3hEakFNQmdOVkJBTU1CV3h2WTJGc2dna0FxZkRYUUgregpQWk13RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUpkbHNtR2k0d1hvSlZ4SnlKVzlDWEZPejhZWkhTWlhicEdsClJXcmI4QkZIMy9SNFhPTTQ5Y3Y4UzErYUZvQ3hJd0taRGFhZVIwNVIxK05jVXNPSnhZL2tGWXNHN3kvMTJFRVIKTi90anVRTGhPNnEzT1piZTUwUFErS2pxbmxURnQrT1ptV1ZQc3NzbUU2WWNhWXBOc09FcmVWZWxNWHNybDBndgpLZ3hLUXFJQ3hJTThNTG1QeUUxSURsS3RSL0RrTkNqNDQybmgvNVlwaUFoOG1BTUE5QjFtcE1uSTZpZFB0RzRhCm5uSGxiVlZaMUE4UWF2bXlCNGRueVZwNlB3QnhSMXJjN0xoV3VBT3V2WkNWWGpxUVZLMkREdEJ0Q3RGcUUrQlQKa1I4MnFjMGtKR2IwYTFkRWRubmJEdU1BQzA0WnRrQnpQTlV6RzNqSkNKVkZhVEsyMVhZQjNkMUR4UTVSUFJyTQptZmVXUlM5cytmdlplOTZmb1BIMmZuREN0aFJkOXI5UXdWSmhZMkRlTzlQOVd0TW9QMWhMempHSnQ1czB2RmpRClE2RDdZaXFHajkvTXhoc1BqZjlucjhCL3ZUVzFWV1hackh2NDBlR01tM1kwRXUvMlA0bFR5YnEyUlp2enhlQTgKTmlXL0lEa3VrbmlMaU95UnZKL1RWVXFsYllhbm9qeUVlbzJaTnp0RTVadTRYUndabnZIeVIzNFVrMldTYjdOSAp4Tk53TUt6ZHdROFpQQnM1OCtZUjByUGlGN2V4WXprVnFOdHNydEYwUjVPRm5kZjN4Yi9GeC8rRTBhd1UyUlZSClNiTzUxM1JIMTFZZ3RneWlMM0kxRjk3NGFDZTRhYkJuWHdmWDA1eUl2a2dEbFAvdXA5T25WRjV6eU8xSkRmUDAKSlFqY20zbGsKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
    "clientKey": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBb1Q5d3FIT2RVRWo2K0ZtaGdsckVhc1BjaXd3MHBLNC8wN3A3NjFWVzdCY1ZOc1lyCnd5dDdQUHhxbWxJODdHNk9UYTM3SlFKT0g5cjFKeVV2VEl6MG1rdnFrME1aa0tCT3Y2U0xEeUNwcDhnNE1HekMKS0dyV05ZQkhGY3Iwd0tLbDVvVXRYbjl0MEc4U1c4bU1GMXFaT2k5aUxkRVl4T2RuR3Jhb1p6OUU2dExUUFNBQwp0Qkc5UFJ0Q3BvRVA5WXl2L1dKMDdnOUx4bUsvc1lVbXBKMHpFZ2xwOWJhVERONTVDcmhQNE5zcUtkdS9LR3dwCis4SVM3MXBncFhLaGFna3kzclhVVnluY2RtcEhLWDRuQTRHSHRLTFIyQTU5UHIrWlZZK2xQV2FwdE54d1hGekYKQ0FRUlJtNWJvUUhFSFNSaE1tcHIvemF6Z1pTMGhubTNJL21pRk5mODkvc0VJalpnQWI1eFp6VFNDMzVLOFpFdwprd0Q5L2F1RUZEd3BvL0RiS3JLM21JNzhwWWp3bE0rMHp3T2tkTmxUdVpRVjZWYVNCUnpMNC9pQXdLZ0NXSU0zCmRvVEZzNmJUdytVeVYrbG4vN2RwMUFlbGtublM1WmQwWG1SSHg1dWVyaFFBRkNjWXc5Sk9uQVk5NW8vUVBqeEUKaFZpL21LZHZnUDZVWDlPWitUdnVQUHpOeWYrTkp4OUVOT3RQNUxWTnBQdVJnMmtNeDZWeG52VHlKOC8wUjUxRgo5bmpndnF3MXQ0NHJnQnkvcTRVaWsrWHZBTDR1VjRpZzJwVStwRWJ5bWp6SjJlM2kzQUpKV0NOYWVHRmt3NmFBCmlOWE5maThUaTVGN214ODBLYnlLQWZWUzA0MmpFSG0wdi9jek03aXE1d1FaRUVJRlA2cWQ5ODNHWms4Q0F3RUEKQVFLQ0FnRUFqUDBpYlRmQjZrd1ZuUWNKOENlYkxGc2JRRDBvM29FNWI5RFR2ejQ4Sld3OWNVb3ZRNVNHU2huTwp3Q1o5L0tEaUxrdWNsNHgvY04wTGsvR3dmTGVXdkQ3NjJVNUhVU3pLRGtrNkNiMGVlb1RYbElmVDhIRVI0Vy9MCk4rUGd3M3F6b203NTczRnVQRnlSNmMyOWYwSUpUbFhWKzRlanA2OUplSk1UaGt0TTRDSDg3NnBJa3RnYjVnMHEKNXRsY2NmQlVoVElNV1lib1U0dE9YMUswS2lVRlhaVDdvQXZHWWU4NFdNWTFtYjhvQzdlSFdqblJMNzlPdlJnQgovMGZPbVI5MzZrR0VhNzQvZFE2U01GYU1tRVV1dWlQUFphR3RveXIyVUZpc081YkRkazkwczEydUxjY1lyOE9ZCnZKd0Z0UkYxSnhia1hSK2dMd0l1SXBMVUxsRjhoR2ZEb3QvWUZvZGJBM3A0dlM1MEc4c212YWlROVhnVXdGdFoKMFdpWU5iQ2JKdlY1d0w2aXlPcTRqakFGeExqU2lqc20ybFpVcTZiVDY5YVFpR0xmbXJDYnJDRnJFSkNpU1F3WgpGenVrMGlXYnM4ZXRySWFnNjdtY0kvTlF4RkVQbXB2ZEdQTllGZ0RwUDdEMnlhUnA4MzVjeWJVd04vcUE2RnBGCkJHSEQ0ZWl6OWd0WlhwT25NYzBibk9KcG5qMlNpaUUwaS95a0NuTlZvRm1SeXladVY0YVlQd29laWpOZ1NjZWIKQTY1TThHaWtUdmdiemFaL3YvOVdlWElURGJVWjcvQTBwYlRDTjZZZ0xscjBpQlJPMUFmeGhNNkVmems4QlFUSQpidWJ5U3ZwQWhiRk83aTZLVVAvZW03eVV0ckxBekdOR2dnTFJMVEpRWGdYUWNSeUdVWkVDZ2dFQkFNOUVhZW5nCm9peE1QaE5jTlN0dENEekhER2RDSFRWcDN2U0s0bkNHRFhGY3N2aytqMGRRM2pQYWx2V3dDMjBYOTZWQlR1bzAKTFhhcGgwZ0xHYloyS3JWQ0tPOWZrZXpQVWMwOWMzNmp1OExFZ3NBZ0dGOUpXeDVlMUVleUIzS0w0RnIreXVVVwpLbDc0azltYmZVK1dRcG1vc3hKdW41ZjBqbzhoclRucGtqR2FqUHhTL05GaE1YVHpGdjl6dEhVblZhTzJVYzI3ClB1WUJNSkxGdTQ2Y2xxWG53RTQxUTVwVHFLZGwxaUw0dFRrVER5ekFNUHZucW9HSld4Q1ZidjR3b2Z1WXBrY0gKSG50SlVEMVhhSnI1N0xoSU4zeDRCN2pFU3JYTUwrOVVIcWgzTDRna2FwV0VTTDlsY1d3aWxlcHhYdG5RdDI3NgprK0NXVVpuY09TUklVM2NDZ2dFQkFNY3BGRGdBWHVYS040b0Z3TC9lRkkvUWVSeXBDakQ2LytidVdCV2F0SXdQCnprQ1h6ODNaY256NmQ4YUlHc2I1bi9tbFYwaHBCdkdtZzBkN0ZHbHZoWDVhZDRzeXFLZWZJajZtWG1PMDRrQ3gKa29kZkErUVh5Q1EwWlFSRi9TenRyNWxWQi9yYnhhYzdCVlplQUxvVUtSSzdUMFhCOHczVXN1SnNyNGR0QVkwZQpDUDlmQlRybFowYjBFSnROWXlVbGFSNWRkaFVuYnJFeHFkUzg5Yk8ybUxBUjVDTVJJRXZScTRseGlQYzRTTXhYCmhUY2FSTmpBU2M3WE5NUUI1WVpXeUxaYzN1RndQaXRHK0sxTGMwNWs1VFpWeEtNdmVhOXBiZkRHLysyQm9IcHIKOEFHVXVKVlk2bCtGTzNleXlmenRPVGtpMUkwNmlxU2dURzVrbzJ0bXlla0NnZ0VBUjVYZWFzdU4xM1RodjdnSwpHUnlJU3MySXFDVTZoMWN3alE5bTAreEl1azJFOXZhM2I2OHJmNGRRdWp4NlJjeVFXTUFzckZFbkhxUEF1STQwCjdFTDF6ekt4aHJOZ2FBVFd3T2NuZTZhN1U3S2hZZy96dXYxUC9qWk1aUkxFNWJnUDNmM0FQODBmQnp3ZGZIdnEKbE5GVjRWSlZ2dGo4UC9SVVJIVWlLaTFVczlNb1BJSEJGZVBXdkFpMWViY1JyYURQUUVMWkVCQkswZys1SWdndgpGanRaQUtZQlVrR3RQcUVFVUFTcEo5ejBZbWtGeGJQL2R4RjFYMVg4WU1icjFka2dLUkI0NVhFOUF1RzRWK2RYCmxxY1pMakNyRVU4M2c0WXdNNGY1U2xTb1hoRUVGcVpWTlp6QnIzRXU4bVVqbUJ4ZDRTYm9JK2xocDZEalFCdkMKbEpoeVV3S0NBUUI4QUpyREo0L3VtVkt0VUZtcjNQV0dlYkgrNDAwaUpCWFRUbEZ2Mml4U0RNRkp2SHc1V2d1TAp2MU4yUEdZWHYzTVl1QmE1VWhOdHdGUjYzQ3BnWDN5SnFJQklIaG1lakZtQkVvc3duMzVEODR3ZFYwNlA1VExMClFBZ3BlZjVoeS9nS2kwUDFzSUxIVmR0RDVER2xxa25Nak8yVnJHWE9GY0h2Y3VaemRxNkJrOUxjVmVobXZGRHEKZjZvYldEckQ5U0FYTlBBQnlkU0U1VHd0NWgxQmNRNXVpaVUycEVJc2t2YXdGQTNJaDdYajdSWlhzYlp1RW9PaQpFcUthNitkaUZvVFA3dEVqSW9UQzQyU1FXYXNJZzQrbm5nMVo0WVJ0Y0VKd3FTYk9WV2g2OE51MTBFaUJUS1JaCkp4Wll0K3hGMjlwR05lYUxySWlJYWZwTXZjSjJhOENKQW9JQkFRRE5Nbk43M1k4UWNwSld4eTk4SmNLSklrTTQKY3BIeUM1U2dva2QvRWdBZzdMdW96a0grQWtWREZTUmcxdmRDSjdSMU9yTXNxeUVyYy9xV0pDYWpqVWJ6UlVyOApJQlN5b2Y1Zko3UHZ1VFc2WkV6RUR2WEw5WW90VExLM1VnUWpGc0N6c3dIUVhOaEE1QVBHWWZZa2hsTnlubVdpCnpyYlFYU1Q2RkxDeSt3UW9UQ3NzVHF4NW8yOEZiZ21RSTZvQzdmQ0ViTzZNL2dCTUxtbjhFNW03dWJVWE9wZmYKaEhsVGcwaVpKU0lDTDNRK25aZ0J4dVNSZ1hsMjJkUjZNYndiRDR6TTltNFhjTjNLTkNGMFY2bkJuQ0dnQ2hNaQp3aWpnS0w5RU1LemZFN0N4K05ZR0x5em9icVBXWDdMak81UjlPUWlpa3FBNmxxTldlN1g0b2NacFZ6c28KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"
  },
  "eu-2": {
    "clientId": "client_id",
    "clientSecret": "client_secret",
    "url": "http://compass-external-services-mock.compass-system.svc.cluster.local:8080",
    "tokenUrl": "https://compass-external-services-mock-sap-mtls.kyma-local:8080",
    "clientCert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUYwakNDQTdxZ0F3SUJBZ0lKQUtudzEwQi9zejJUTUEwR0NTcUdTSWIzRFFFQkN3VUFNRTB4Q3pBSkJnTlYKQkFZVEFrSkhNUTR3REFZRFZRUUlEQVZzYjJOaGJERU9NQXdHQTFVRUJ3d0ZiRzlqWVd3eERqQU1CZ05WQkFvTQpCV3h2WTJGc01RNHdEQVlEVlFRRERBVnNiMk5oYkRBZ0Z3MHlNakF5TWpVeE16VTFNVEJhR0E4eU1USXlNREl3Ck1URXpOVFV4TUZvd1RURUxNQWtHQTFVRUJoTUNRa2N4RGpBTUJnTlZCQWdNQld4dlkyRnNNUTR3REFZRFZRUUgKREFWc2IyTmhiREVPTUF3R0ExVUVDZ3dGYkc5allXd3hEakFNQmdOVkJBTU1CV3h2WTJGc01JSUNJakFOQmdrcQpoa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQW9UOXdxSE9kVUVqNitGbWhnbHJFYXNQY2l3dzBwSzQvCjA3cDc2MVZXN0JjVk5zWXJ3eXQ3UFB4cW1sSTg3RzZPVGEzN0pRSk9IOXIxSnlVdlRJejBta3ZxazBNWmtLQk8KdjZTTER5Q3BwOGc0TUd6Q0tHcldOWUJIRmNyMHdLS2w1b1V0WG45dDBHOFNXOG1NRjFxWk9pOWlMZEVZeE9kbgpHcmFvWno5RTZ0TFRQU0FDdEJHOVBSdENwb0VQOVl5di9XSjA3ZzlMeG1LL3NZVW1wSjB6RWdscDliYVRETjU1CkNyaFA0TnNxS2R1L0tHd3ArOElTNzFwZ3BYS2hhZ2t5M3JYVVZ5bmNkbXBIS1g0bkE0R0h0S0xSMkE1OVByK1oKVlkrbFBXYXB0Tnh3WEZ6RkNBUVJSbTVib1FIRUhTUmhNbXByL3phemdaUzBobm0zSS9taUZOZjg5L3NFSWpaZwpBYjV4WnpUU0MzNUs4WkV3a3dEOS9hdUVGRHdwby9EYktySzNtSTc4cFlqd2xNKzB6d09rZE5sVHVaUVY2VmFTCkJSekw0L2lBd0tnQ1dJTTNkb1RGczZiVHcrVXlWK2xuLzdkcDFBZWxrbm5TNVpkMFhtUkh4NXVlcmhRQUZDY1kKdzlKT25BWTk1by9RUGp4RWhWaS9tS2R2Z1A2VVg5T1orVHZ1UFB6TnlmK05KeDlFTk90UDVMVk5wUHVSZzJrTQp4NlZ4bnZUeUo4LzBSNTFGOW5qZ3ZxdzF0NDRyZ0J5L3E0VWlrK1h2QUw0dVY0aWcycFUrcEVieW1qekoyZTNpCjNBSkpXQ05hZUdGa3c2YUFpTlhOZmk4VGk1RjdteDgwS2J5S0FmVlMwNDJqRUhtMHYvY3pNN2lxNXdRWkVFSUYKUDZxZDk4M0daazhDQXdFQUFhT0JzakNCcnpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJUbwprMzBXVzhCRUxnMEZBWEZscjNnYnpkektvVEI5QmdOVkhTTUVkakIwZ0JUb2szMFdXOEJFTGcwRkFYRmxyM2diCnpkektvYUZScEU4d1RURUxNQWtHQTFVRUJoTUNRa2N4RGpBTUJnTlZCQWdNQld4dlkyRnNNUTR3REFZRFZRUUgKREFWc2IyTmhiREVPTUF3R0ExVUVDZ3dGYkc5allXd3hEakFNQmdOVkJBTU1CV3h2WTJGc2dna0FxZkRYUUgregpQWk13RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUpkbHNtR2k0d1hvSlZ4SnlKVzlDWEZPejhZWkhTWlhicEdsClJXcmI4QkZIMy9SNFhPTTQ5Y3Y4UzErYUZvQ3hJd0taRGFhZVIwNVIxK05jVXNPSnhZL2tGWXNHN3kvMTJFRVIKTi90anVRTGhPNnEzT1piZTUwUFErS2pxbmxURnQrT1ptV1ZQc3NzbUU2WWNhWXBOc09FcmVWZWxNWHNybDBndgpLZ3hLUXFJQ3hJTThNTG1QeUUxSURsS3RSL0RrTkNqNDQybmgvNVlwaUFoOG1BTUE5QjFtcE1uSTZpZFB0RzRhCm5uSGxiVlZaMUE4UWF2bXlCNGRueVZwNlB3QnhSMXJjN0xoV3VBT3V2WkNWWGpxUVZLMkREdEJ0Q3RGcUUrQlQKa1I4MnFjMGtKR2IwYTFkRWRubmJEdU1BQzA0WnRrQnpQTlV6RzNqSkNKVkZhVEsyMVhZQjNkMUR4UTVSUFJyTQptZmVXUlM5cytmdlplOTZmb1BIMmZuREN0aFJkOXI5UXdWSmhZMkRlTzlQOVd0TW9QMWhMempHSnQ1czB2RmpRClE2RDdZaXFHajkvTXhoc1BqZjlucjhCL3ZUVzFWV1hackh2NDBlR01tM1kwRXUvMlA0bFR5YnEyUlp2enhlQTgKTmlXL0lEa3VrbmlMaU95UnZKL1RWVXFsYllhbm9qeUVlbzJaTnp0RTVadTRYUndabnZIeVIzNFVrMldTYjdOSAp4Tk53TUt6ZHdROFpQQnM1OCtZUjByUGlGN2V4WXprVnFOdHNydEYwUjVPRm5kZjN4Yi9GeC8rRTBhd1UyUlZSClNiTzUxM1JIMTFZZ3RneWlMM0kxRjk3NGFDZTRhYkJuWHdmWDA1eUl2a2dEbFAvdXA5T25WRjV6eU8xSkRmUDAKSlFqY20zbGsKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
    "clientKey": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBb1Q5d3FIT2RVRWo2K0ZtaGdsckVhc1BjaXd3MHBLNC8wN3A3NjFWVzdCY1ZOc1lyCnd5dDdQUHhxbWxJODdHNk9UYTM3SlFKT0g5cjFKeVV2VEl6MG1rdnFrME1aa0tCT3Y2U0xEeUNwcDhnNE1HekMKS0dyV05ZQkhGY3Iwd0tLbDVvVXRYbjl0MEc4U1c4bU1GMXFaT2k5aUxkRVl4T2RuR3Jhb1p6OUU2dExUUFNBQwp0Qkc5UFJ0Q3BvRVA5WXl2L1dKMDdnOUx4bUsvc1lVbXBKMHpFZ2xwOWJhVERONTVDcmhQNE5zcUtkdS9LR3dwCis4SVM3MXBncFhLaGFna3kzclhVVnluY2RtcEhLWDRuQTRHSHRLTFIyQTU5UHIrWlZZK2xQV2FwdE54d1hGekYKQ0FRUlJtNWJvUUhFSFNSaE1tcHIvemF6Z1pTMGhubTNJL21pRk5mODkvc0VJalpnQWI1eFp6VFNDMzVLOFpFdwprd0Q5L2F1RUZEd3BvL0RiS3JLM21JNzhwWWp3bE0rMHp3T2tkTmxUdVpRVjZWYVNCUnpMNC9pQXdLZ0NXSU0zCmRvVEZzNmJUdytVeVYrbG4vN2RwMUFlbGtublM1WmQwWG1SSHg1dWVyaFFBRkNjWXc5Sk9uQVk5NW8vUVBqeEUKaFZpL21LZHZnUDZVWDlPWitUdnVQUHpOeWYrTkp4OUVOT3RQNUxWTnBQdVJnMmtNeDZWeG52VHlKOC8wUjUxRgo5bmpndnF3MXQ0NHJnQnkvcTRVaWsrWHZBTDR1VjRpZzJwVStwRWJ5bWp6SjJlM2kzQUpKV0NOYWVHRmt3NmFBCmlOWE5maThUaTVGN214ODBLYnlLQWZWUzA0MmpFSG0wdi9jek03aXE1d1FaRUVJRlA2cWQ5ODNHWms4Q0F3RUEKQVFLQ0FnRUFqUDBpYlRmQjZrd1ZuUWNKOENlYkxGc2JRRDBvM29FNWI5RFR2ejQ4Sld3OWNVb3ZRNVNHU2huTwp3Q1o5L0tEaUxrdWNsNHgvY04wTGsvR3dmTGVXdkQ3NjJVNUhVU3pLRGtrNkNiMGVlb1RYbElmVDhIRVI0Vy9MCk4rUGd3M3F6b203NTczRnVQRnlSNmMyOWYwSUpUbFhWKzRlanA2OUplSk1UaGt0TTRDSDg3NnBJa3RnYjVnMHEKNXRsY2NmQlVoVElNV1lib1U0dE9YMUswS2lVRlhaVDdvQXZHWWU4NFdNWTFtYjhvQzdlSFdqblJMNzlPdlJnQgovMGZPbVI5MzZrR0VhNzQvZFE2U01GYU1tRVV1dWlQUFphR3RveXIyVUZpc081YkRkazkwczEydUxjY1lyOE9ZCnZKd0Z0UkYxSnhia1hSK2dMd0l1SXBMVUxsRjhoR2ZEb3QvWUZvZGJBM3A0dlM1MEc4c212YWlROVhnVXdGdFoKMFdpWU5iQ2JKdlY1d0w2aXlPcTRqakFGeExqU2lqc20ybFpVcTZiVDY5YVFpR0xmbXJDYnJDRnJFSkNpU1F3WgpGenVrMGlXYnM4ZXRySWFnNjdtY0kvTlF4RkVQbXB2ZEdQTllGZ0RwUDdEMnlhUnA4MzVjeWJVd04vcUE2RnBGCkJHSEQ0ZWl6OWd0WlhwT25NYzBibk9KcG5qMlNpaUUwaS95a0NuTlZvRm1SeXladVY0YVlQd29laWpOZ1NjZWIKQTY1TThHaWtUdmdiemFaL3YvOVdlWElURGJVWjcvQTBwYlRDTjZZZ0xscjBpQlJPMUFmeGhNNkVmems4QlFUSQpidWJ5U3ZwQWhiRk83aTZLVVAvZW03eVV0ckxBekdOR2dnTFJMVEpRWGdYUWNSeUdVWkVDZ2dFQkFNOUVhZW5nCm9peE1QaE5jTlN0dENEekhER2RDSFRWcDN2U0s0bkNHRFhGY3N2aytqMGRRM2pQYWx2V3dDMjBYOTZWQlR1bzAKTFhhcGgwZ0xHYloyS3JWQ0tPOWZrZXpQVWMwOWMzNmp1OExFZ3NBZ0dGOUpXeDVlMUVleUIzS0w0RnIreXVVVwpLbDc0azltYmZVK1dRcG1vc3hKdW41ZjBqbzhoclRucGtqR2FqUHhTL05GaE1YVHpGdjl6dEhVblZhTzJVYzI3ClB1WUJNSkxGdTQ2Y2xxWG53RTQxUTVwVHFLZGwxaUw0dFRrVER5ekFNUHZucW9HSld4Q1ZidjR3b2Z1WXBrY0gKSG50SlVEMVhhSnI1N0xoSU4zeDRCN2pFU3JYTUwrOVVIcWgzTDRna2FwV0VTTDlsY1d3aWxlcHhYdG5RdDI3NgprK0NXVVpuY09TUklVM2NDZ2dFQkFNY3BGRGdBWHVYS040b0Z3TC9lRkkvUWVSeXBDakQ2LytidVdCV2F0SXdQCnprQ1h6ODNaY256NmQ4YUlHc2I1bi9tbFYwaHBCdkdtZzBkN0ZHbHZoWDVhZDRzeXFLZWZJajZtWG1PMDRrQ3gKa29kZkErUVh5Q1EwWlFSRi9TenRyNWxWQi9yYnhhYzdCVlplQUxvVUtSSzdUMFhCOHczVXN1SnNyNGR0QVkwZQpDUDlmQlRybFowYjBFSnROWXlVbGFSNWRkaFVuYnJFeHFkUzg5Yk8ybUxBUjVDTVJJRXZScTRseGlQYzRTTXhYCmhUY2FSTmpBU2M3WE5NUUI1WVpXeUxaYzN1RndQaXRHK0sxTGMwNWs1VFpWeEtNdmVhOXBiZkRHLysyQm9IcHIKOEFHVXVKVlk2bCtGTzNleXlmenRPVGtpMUkwNmlxU2dURzVrbzJ0bXlla0NnZ0VBUjVYZWFzdU4xM1RodjdnSwpHUnlJU3MySXFDVTZoMWN3alE5bTAreEl1azJFOXZhM2I2OHJmNGRRdWp4NlJjeVFXTUFzckZFbkhxUEF1STQwCjdFTDF6ekt4aHJOZ2FBVFd3T2NuZTZhN1U3S2hZZy96dXYxUC9qWk1aUkxFNWJnUDNmM0FQODBmQnp3ZGZIdnEKbE5GVjRWSlZ2dGo4UC9SVVJIVWlLaTFVczlNb1BJSEJGZVBXdkFpMWViY1JyYURQUUVMWkVCQkswZys1SWdndgpGanRaQUtZQlVrR3RQcUVFVUFTcEo5ejBZbWtGeGJQL2R4RjFYMVg4WU1icjFka2dLUkI0NVhFOUF1RzRWK2RYCmxxY1pMakNyRVU4M2c0WXdNNGY1U2xTb1hoRUVGcVpWTlp6QnIzRXU4bVVqbUJ4ZDRTYm9JK2xocDZEalFCdkMKbEpoeVV3S0NBUUI4QUpyREo0L3VtVkt0VUZtcjNQV0dlYkgrNDAwaUpCWFRUbEZ2Mml4U0RNRkp2SHc1V2d1TAp2MU4yUEdZWHYzTVl1QmE1VWhOdHdGUjYzQ3BnWDN5SnFJQklIaG1lakZtQkVvc3duMzVEODR3ZFYwNlA1VExMClFBZ3BlZjVoeS9nS2kwUDFzSUxIVmR0RDVER2xxa25Nak8yVnJHWE9GY0h2Y3VaemRxNkJrOUxjVmVobXZGRHEKZjZvYldEckQ5U0FYTlBBQnlkU0U1VHd0NWgxQmNRNXVpaVUycEVJc2t2YXdGQTNJaDdYajdSWlhzYlp1RW9PaQpFcUthNitkaUZvVFA3dEVqSW9UQzQyU1FXYXNJZzQrbm5nMVo0WVJ0Y0VKd3FTYk9WV2g2OE51MTBFaUJUS1JaCkp4Wll0K3hGMjlwR05lYUxySWlJYWZwTXZjSjJhOENKQW9JQkFRRE5Nbk43M1k4UWNwSld4eTk4SmNLSklrTTQKY3BIeUM1U2dva2QvRWdBZzdMdW96a0grQWtWREZTUmcxdmRDSjdSMU9yTXNxeUVyYy9xV0pDYWpqVWJ6UlVyOApJQlN5b2Y1Zko3UHZ1VFc2WkV6RUR2WEw5WW90VExLM1VnUWpGc0N6c3dIUVhOaEE1QVBHWWZZa2hsTnlubVdpCnpyYlFYU1Q2RkxDeSt3UW9UQ3NzVHF4NW8yOEZiZ21RSTZvQzdmQ0ViTzZNL2dCTUxtbjhFNW03dWJVWE9wZmYKaEhsVGcwaVpKU0lDTDNRK25aZ0J4dVNSZ1hsMjJkUjZNYndiRDR6TTltNFhjTjNLTkNGMFY2bkJuQ0dnQ2hNaQp3aWpnS0w5RU1LemZFN0N4K05ZR0x5em9icVBXWDdMak81UjlPUWlpa3FBNmxxTldlN1g0b2NacFZ6c28KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"
  }
}
EOF

# This file contains regional dependencies configuration for tenants service
cat <<EOF > /tmp/dependencies.json
    {
        "eu-1": "{\n \"xsappname\": \"xsappname1\", \"clientid\": \"clientid-1\", \"certificate\": \"client-cert-1\", \"key\": \"client-cert-key-1\", \"url\": \"http://token-url\", \"uri\": \"http://destination-service\"\n}",
        "eu-2": "{\n \"xsappname\": \"xsappname2\", \"clientid\": \"clientid-2\", \"certificate\": \"client-cert-2\", \"key\": \"client-cert-key-2\", \"url\": \"http://token-url\", \"uri\": \"http://destination-service\"\n}"
    }
EOF

kubectl create secret generic "$CLIENT_CERT_SECRET_NAME" --from-literal="$APP_EXTERNAL_CLIENT_CERT_KEY"="$APP_EXTERNAL_CLIENT_CERT_VALUE" --from-literal="$APP_EXTERNAL_CLIENT_KEY_KEY"="$APP_EXTERNAL_CLIENT_KEY_VALUE" --save-config --dry-run=client -o yaml | kubectl apply -f -

# pairing adapters configmap needed for the watcher started in the director
kubectl create configmap "$APP_PAIRING_ADAPTER_CM_NAME" --from-literal="$APP_PAIRING_ADAPTER_CM_KEY"='{"d3e9b9f5-25dc-4adb-a0a0-ed69ef371fb6":"http://compass-pairing-adapter.compass-system.svc.cluster.local/adapter-local-mtls"}'

if [[  ${DEBUG} == true ]]; then
    echo -e "${GREEN}Debug mode activated on port $DEBUG_PORT${NC}"
    cd $GOPATH/src/github.com/kyma-incubator/compass/components/director
    CGO_ENABLED=0 go build -gcflags="all=-N -l" ./cmd/${COMPONENT}
    dlv --listen=:$DEBUG_PORT --headless=true --api-version=2 exec ./${COMPONENT}
else
    if [[  ${AUTO_TERMINATE} == true ]]; then
        cd ${ROOT_PATH}
        go build ${ROOT_PATH}/cmd/${COMPONENT}/main.go
        MAIN_APP_LOGFILE=${ROOT_PATH}/main.log

        ${ROOT_PATH}/main > ${MAIN_APP_LOGFILE} &
        MAIN_PROCESS_PID="$!"

        START_TIME=$(date +%s)
        SECONDS=0
        while (( SECONDS < ${TERMINAION_TIMEOUT_IN_SECONDS} )) ; do
            CURRENT_TIME=$(date +%s)
            SECONDS=$((CURRENT_TIME-START_TIME))
            SECONDS_LEFT=$((TERMINAION_TIMEOUT_IN_SECONDS-SECONDS))
            echo "[Director] left ${SECONDS_LEFT} seconds. Wait ..."
            sleep 10
        done

        echo "Timeout of ${TERMINAION_TIMEOUT_IN_SECONDS} seconds for starting director reached. Killing the process."
        echo -e "${GREEN}Kill main process..${NC}"
        kill -SIGINT "${MAIN_PROCESS_PID}"
        echo -e "${GREEN}Delete build result ...${NC}"
        rm ${ROOT_PATH}/main || true
        wait
    else
        go run ${ROOT_PATH}/cmd/${COMPONENT}/main.go
    fi
fi
