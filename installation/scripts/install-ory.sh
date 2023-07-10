#!/usr/bin/env bash

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

# Copy the IDP configuration from the Compass chart; the Compass values are changed by the `run.sh`
VALUES_FILE_COMPASS="${ROOT_PATH}"/chart/compass/values.yaml
IDP_HOST=$(yq ".global.cockpit.auth.idpHost" $VALUES_FILE_COMPASS)
AUTH_PATH=$(yq ".global.cockpit.auth.path" $VALUES_FILE_COMPASS)

VALUES_FILE_ORY="${ROOT_PATH}"/chart/ory/values.yaml
OVERRIDE_TEMP_ORY=ory-temp-values.yaml

function cleanup_trap(){
  if [[ -f "$OVERRIDE_TEMP_ORY" ]]; then
    rm -f "$OVERRIDE_TEMP_ORY"
  fi
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

echo "Values are: $VALUES_FILE_ORY"
echo "Helm install ORY components"
helm upgrade --install ory -f "${VALUES_FILE_ORY}" -n kyma-system "${ROOT_PATH}"/chart/ory --timeout 10m

kubectl set image -n kyma-system cronjob/oathkeeper-jwks-rotator keys-generator=oryd/oathkeeper:v0.38.23
kubectl patch cronjob -n kyma-system oathkeeper-jwks-rotator -p '{"spec":{"schedule": "*/1 * * * *"}}'
until [[ $(kubectl get cronjob -n kyma-system oathkeeper-jwks-rotator --output=jsonpath={.status.lastScheduleTime}) ]]; do
    echo "Waiting for cronjob oathkeeper-jwks-rotator to be scheduled"
    sleep 3
done
kubectl patch cronjob -n kyma-system oathkeeper-jwks-rotator -p '{"spec":{"schedule": "0 0 1 * *"}}'