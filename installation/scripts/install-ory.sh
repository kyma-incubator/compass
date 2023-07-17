#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

# Copy the IDP configuration from the Compass chart; the Compass values are changed by the `run.sh`
VALUES_FILE_COMPASS="${ROOT_PATH}"/chart/compass/values.yaml
IDP_HOST=$(yq ".global.cockpit.auth.idpHost" $VALUES_FILE_COMPASS)
AUTH_PATH=$(yq ".global.cockpit.auth.path" $VALUES_FILE_COMPASS)

VALUES_FILE_ORY="${ROOT_PATH}"/chart/ory/values.yaml
OVERRIDE_TEMP_ORY=ory-temp-values.yaml

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

# If IDP has been configured override the default values in the ORY chart values
if [ ! -z "$IDP_HOST" ]; then
  JWKS_URL=$IDP_HOST$AUTH_PATH

  cp "$VALUES_FILE_ORY" "$OVERRIDE_TEMP_ORY"

  # Overwrite the jwks_url of the Ory Oathkeeper to match the Compass values
  yq -i ".oathkeeper.oathkeeper.config.authenticators.jwt.config.jwks_urls = [\"$JWKS_URL\"]" "${OVERRIDE_TEMP_ORY}"
  
  # Use the temp ORY Helm values ofr the installation if IDP was provided
  VALUES_FILE_ORY="$OVERRIDE_TEMP_ORY"
fi

trap cleanup_trap EXIT INT TERM

echo "Helm install ORY components..."
RELEASE_NS=kyma-system
SECRET_NAME=ory-hydra-credentials
if [[ ! $(kubectl get secret $SECRET_NAME -n kyma-system) ]]; then
  VALUES_DIR="${ROOT_PATH}"/chart/ory/values.yaml

  POSTGRES_USERNAME=$(yq .global.postgresql.postgresqlUsername ${VALUES_DIR})
  POSTGRES_PASSWORD=$(generate_random 10)
  POSTGRES_DB=$(yq .global.postgresql.postgresqlDatabase ${VALUES_DIR})

  DSN=postgres://${POSTGRES_USERNAME}:${POSTGRES_PASSWORD}@ory-postgresql.${RELEASE_NS}.svc.cluster.local:5432/${POSTGRES_DB}?sslmode=disable\&max_conn_lifetime=10s

  SYSTEM=$(generate_random 32)
  COOKIE=$(generate_random 32)
  
  SECRET=$(cat <<EOF
  apiVersion: v1
  kind: Secret
  metadata:
    name: ${SECRET_NAME}
    namespace: ${RELEASE_NS}
  type: Opaque
  data:
    dsn: $(echo -n "${DSN}" | base64)
    secretsSystem: $(echo -n "${SYSTEM}" | base64)
    secretsCookie: $(echo -n "${COOKIE}" | base64)
    postgresql-password: $(echo -n "${POSTGRES_PASSWORD}" | base64)
EOF
)

  echo "Applying database secret"
  set -e
  echo "${SECRET}" | kubectl apply -f -
fi

# --wait is excluded as the deployment hangs; it hangs as there is a cronjob that rotates the jwks secret
helm upgrade --install ory -f "${VALUES_FILE_ORY}" -n $RELEASE_NS "${ROOT_PATH}"/chart/ory

kubectl set image -n kyma-system cronjob/oathkeeper-jwks-rotator keys-generator=oryd/oathkeeper:v0.38.23
kubectl patch cronjob -n kyma-system oathkeeper-jwks-rotator -p '{"spec":{"schedule": "*/1 * * * *"}}'
until [[ $(kubectl get cronjob -n kyma-system oathkeeper-jwks-rotator --output=jsonpath={.status.lastScheduleTime}) ]]; do
    echo "Waiting for cronjob oathkeeper-jwks-rotator to be scheduled"
    sleep 3
done
kubectl patch cronjob -n kyma-system oathkeeper-jwks-rotator -p '{"spec":{"schedule": "0 0 1 * *"}}'