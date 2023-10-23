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

# Copy the identity provider configuration from the Compass chart; the Compass values are changed by the `run.sh`
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

# As of Kyma 2.6.3 we need to specify which namespaces should enable istio injection
kubectl create ns $RELEASE_NS --dry-run=client -o yaml | kubectl apply -f -
kubectl label ns $RELEASE_NS istio-injection=enabled --overwrite
# As of Kubernetes 1.25 we need to replace PodSecurityPolicies; we chose the Pod Security Standards
kubectl label ns $RELEASE_NS pod-security.kubernetes.io/enforce=baseline --overwrite

CLOUD_PERSISTENCE=$(yq ".global.ory.hydra.persistence.gcloud.enabled" ${OVERRIDE_TEMP_ORY})
# The System secret and cookie secret, needed by Hydra, are created by the Secret component of the Helm chart
# Rotating the secrets has to be done manually; the rotation of the Hydra Secrets should be done following this guide: https://www.ory.sh/docs/hydra/self-hosted/secrets-key-rotation
# Hydra requires data persistence, locally the postgres DB of compass is used.
# The connection string(DSN) has to be created
if [ "$CLOUD_PERSISTENCE" = false ]; then
  # Hydra uses the `localdb` instance as its persistence backend
  VALUES_FILE_DB="${ROOT_PATH}"/chart/localdb/values.yaml
  POSTGRES_USERNAME=$(yq ".postgresql.postgresqlUsername" $VALUES_FILE_DB)
  POSTGRES_PASSWORD=$(yq ".postgresql.postgresqlPassword" $VALUES_FILE_DB)
  POSTGRES_DB=$(yq ".global.database.embedded.hydra.name" $VALUES_FILE_DB)

  DSN=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@compass-postgresql.compass-system.svc.cluster.local:5432/${POSTGRES_DB}?sslmode=disable\&max_conn_lifetime=10s

  yq -i ".hydra.hydra.config.dsn = \"$DSN\"" "${OVERRIDE_TEMP_ORY}"
fi

helm upgrade --atomic --install --create-namespace --timeout "${TIMEOUT}" $RELEASE_NAME -f "${OVERRIDE_TEMP_ORY}" -n $RELEASE_NS "${ROOT_PATH}"/chart/ory
