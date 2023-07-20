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
RELEASE_NS=ory
SECRET_NAME=ory-hydra-credentials

kubectl create ns $RELEASE_NS || true

# Create Secret that is referenced as 'existingSecret' under chart/ory/values.yaml
# Secret should not be recreated if it exists, mainly during Helm updates, as it will create new random values.
# The new random values will triggered the redeployment of the postgres db and Hydra - that breaks the deployment
# Rotating the secrets has to be done manually; the rotation of the Hydra Secrets should be done following this guide: https://www.ory.sh/docs/hydra/self-hosted/secrets-key-rotation
if [[ ! $(kubectl get secret $SECRET_NAME -n ${RELEASE_NS}) ]]; then
  echo "Creating secret to be used by the Ory Helm Chart..."
  POSTGRES_USERNAME=$(yq .global.postgresql.postgresqlUsername ${VALUES_FILE_ORY})
  POSTGRES_PASSWORD=$(generate_random 10)
  POSTGRES_DB=$(yq .global.postgresql.postgresqlDatabase ${VALUES_FILE_ORY})

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

RELEASE_NAME=ory

# --wait is excluded as the deployment hangs; it hangs as there is a cronjob that creates the jwks secret
helm upgrade --install $RELEASE_NAME -f "${VALUES_FILE_ORY}" -n $RELEASE_NS "${ROOT_PATH}"/chart/ory

CRONJOB=oathkeeper-jwks-rotator
# CronJob creates a Secret that is needed for the successful deployment of Oathkeeper
kubectl set image -n $RELEASE_NS cronjob $CRONJOB keys-generator=oryd/oathkeeper:v0.38.23
kubectl patch cronjob -n $RELEASE_NS $CRONJOB -p '{"spec":{"schedule": "*/1 * * * *"}}'
until [[ $(kubectl get cronjob -n $RELEASE_NS $CRONJOB --output=jsonpath={.status.lastScheduleTime}) ]]; do
    echo "Waiting for cronjob $CRONJOB to be scheduled"
    sleep 3
done
kubectl patch cronjob -n $RELEASE_NS $CRONJOB -p '{"spec":{"schedule": "0 0 1 * *"}}'

echo "Waiting for oathkeeper deployment to roll out..."
# Wait for Oathkeeper deployment to roll out as the CronJob created the Secret; needs to be ready for successful compass installation
if [[ ! $(kubectl rollout status deployment $RELEASE_NAME-oathkeeper -n $RELEASE_NS --timeout=30m) ]]; then
  echo "Oathkeeper did not deploy correctly..."
  echo "Uninstalling Ory Helm chart..."
  helm uninstall $RELEASE_NAME -n $RELEASE_NS
  exit 1
fi