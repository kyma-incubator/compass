#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

VALUES_FILE_ORY="${ROOT_PATH}"/chart/ory/values.yaml
OVERRIDE_TEMP_ORY=ory-temp-values.yaml

TIMEOUT=30m

# Always create temporary override file, which is used whether "--overrides-file" is used or not
cp "$VALUES_FILE_ORY" "$OVERRIDE_TEMP_ORY"

# checkInputParameterValue is a function to check if input parameter is valid
# There HAS to be provided argument:
# $1 - value for input parameter
# for example in installation/cmd/run.sh we can set --vm-driver argument, which has to have a value.

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --overrides-file)
            checkInputParameterValue "${2}"
            yq eval-all --inplace '. as $item ireduce ({}; . * $item )' ${OVERRIDE_TEMP_ORY} ${2}
            shift # past argument
            shift
        ;;
        --timeout)
            checkInputParameterValue "${2}"
            TIMEOUT="${2}"
            shift # past argument
            shift
        ;;
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

# Remove the temporary Ory values.yaml file created
function cleanup_trap(){
  if [[ -f "$OVERRIDE_TEMP_ORY" ]]; then
    rm -f "$OVERRIDE_TEMP_ORY"
  fi
}

# Generate random string with the length given as argument
function generate_random(){
  cat /dev/urandom | LC_ALL=C tr -dc 'a-z0-9' | fold -w ${1} | head -n 1
}

# Copy the IDP configuration from the Compass chart; the Compass values are changed by the `run.sh`
VALUES_FILE_COMPASS="${ROOT_PATH}"/chart/compass/values.yaml
IDP_HOST=$(yq ".global.cockpit.auth.idpHost" $VALUES_FILE_COMPASS)
AUTH_PATH=$(yq ".global.cockpit.auth.path" $VALUES_FILE_COMPASS)

# If IDP has been configured override the default values in the ORY chart values
if [ ! -z "$IDP_HOST" ]; then
  JWKS_URL=$IDP_HOST$AUTH_PATH

  # Overwrite the jwks_url of the Ory Oathkeeper to match the Compass values
  yq -i ".oathkeeper.oathkeeper.config.authenticators.jwt.config.jwks_urls = [\"$JWKS_URL\"]" "${OVERRIDE_TEMP_ORY}"
fi

trap cleanup_trap EXIT INT TERM

echo "Helm install ORY components..."
RELEASE_NS=ory
RELEASE_NAME=ory-stack
SECRET_NAME=ory-hydra-credentials

kubectl create ns $RELEASE_NS || true

LOCAL_PERSISTENCE=$(yq ".global.ory.hydra.persistence.postgresql.enabled" ${OVERRIDE_TEMP_ORY})

# Create Secret that is referenced as 'existingSecret' under chart/ory/values.yaml
# Secret should not be recreated if it exists, mainly during Helm updates, as it will create new random values.
# The new random values will triggered the redeployment of the postgres db and Hydra - that breaks the deployment
# Rotating the secrets has to be done manually; the rotation of the Hydra Secrets should be done following this guide: https://www.ory.sh/docs/hydra/self-hosted/secrets-key-rotation
if [ ! "$(kubectl get secret $SECRET_NAME -n ${RELEASE_NS})" -a "$LOCAL_PERSISTENCE" = true ]; then
  echo "Creating secret to be used by the Ory Helm Chart..."
  POSTGRES_USERNAME=$(yq .global.postgresql.postgresqlUsername ${OVERRIDE_TEMP_ORY})
  POSTGRES_PASSWORD=$(generate_random 10)
  POSTGRES_DB=$(yq .global.postgresql.postgresqlDatabase ${OVERRIDE_TEMP_ORY})

  DSN=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@${RELEASE_NAME}-postgresql.${RELEASE_NS}.svc.cluster.local:5432/${POSTGRES_DB}?sslmode=disable\&max_conn_lifetime=10s

  SYSTEM=$(generate_random 32)
  COOKIE=$(generate_random 32)

  echo "Creating Ory credentials Secret"
  kubectl create secret generic "$SECRET_NAME" -n "$RELEASE_NS" \
    --from-literal=dsn="${DSN}" \
    --from-literal=secretsSystem="${SYSTEM}" \
    --from-literal=secretsCookie="${COOKIE}" \
    --from-literal=postgresql-password="${POSTGRES_PASSWORD}" \
    --dry-run=client -o yaml | kubectl apply -f -
fi

# --wait is excluded as the deployment hangs; it hangs as there is a cronjob that creates the jwks secret for Oathkeeper
# This cronjob is triggered in the statements below
helm upgrade --install $RELEASE_NAME -f "${OVERRIDE_TEMP_ORY}" -n $RELEASE_NS "${ROOT_PATH}"/chart/ory

CRONJOB=oathkeeper-jwks-rotator
# CronJob creates a Secret that is needed for the successful deployment of Oathkeeper
kubectl set image -n $RELEASE_NS cronjob $CRONJOB keys-generator=oryd/oathkeeper:v0.38.23
kubectl patch cronjob -n $RELEASE_NS $CRONJOB -p '{"spec":{"schedule": "*/1 * * * *"}}'
until [[ $(kubectl get cronjob -n $RELEASE_NS $CRONJOB --output=jsonpath={.status.lastScheduleTime}) ]]; do
    echo "Waiting for cronjob $CRONJOB to be scheduled"
    sleep 3
done
kubectl patch cronjob -n $RELEASE_NS $CRONJOB -p '{"spec":{"schedule": "0 0 1 * *"}}'

RESULT=0
PIDS=""

kubectl rollout status deployment $RELEASE_NAME-hydra -n $RELEASE_NS --timeout=$TIMEOUT &
PIDS="$PIDS $!"

kubectl rollout status deployment $RELEASE_NAME-oathkeeper -n $RELEASE_NS --timeout=$TIMEOUT &
PIDS="$PIDS $!"

# Wait for Ory deployment to roll out as they needs to be ready for successful compass installation
for PID in $PIDS; do
  wait $PID || let "RESULT=1"
done

if [ "$RESULT" == "1" ]; then
  echo "Ory components did not deploy correctly..."
  echo "Uninstalling Ory Helm chart and removing namespace"
  helm uninstall $RELEASE_NAME -n $RELEASE_NS
  kubectl delete ns $RELEASE_NS
  exit 1
fi
