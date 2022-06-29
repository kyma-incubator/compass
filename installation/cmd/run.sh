#!/usr/bin/env bash

set -o errexit

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh
source $SCRIPTS_DIR/prom-mtls-patch.sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..
PATH_TO_VALUES="$CURRENT_DIR/../../chart/compass/values.yaml"
PATH_TO_HYDRATOR_VALUES="$CURRENT_DIR/../../chart/compass/charts/hydrator/values.yaml"
PATH_TO_COMPASS_OIDC_CONFIG_FILE="$HOME/.compass.yaml"
MIGRATOR_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/migrator-job.yaml)
UPDATE_EXPECTED_SCHEMA_VERSION_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/update-expected-schema-version-job.yaml)
SCHEMA_MIGRATOR_COMPONENT_PATH=${ROOT_PATH}/components/schema-migrator
RESET_VALUES_YAML=true

K3D_MEMORY=8192MB
K3D_TIMEOUT=10m0s
APISERVER_VERSION=1.20.11

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --kyma-release)
            checkInputParameterValue "${2}"
            KYMA_RELEASE="${2}"
            shift # past argument
        ;;
        --kyma-installation)
            checkInputParameterValue "${2}"
            KYMA_INSTALLATION="${2}"
            shift # past argument
            shift # past value
        ;;
        --skip-k3d-start)
            SKIP_K3D_START=true
            shift # past argument
        ;;
        --skip-kyma-start)
            SKIP_KYMA_START=true
            shift # past argument
        ;;
        --dump-db)
            DUMP_DB=true
            DUMP_IMAGE_TAG="dump"
            shift # past argument
        ;;
        --k3d-memory)
            checkInputParameterValue "${2}"
            K3D_MEMORY="${2}"
            shift # past argument
            shift # past value
        ;;
        --k3d-timeout)
            checkInputParameterValue "${2}"
            K3D_TIMEOUT="${2}"
            shift # past argument
            shift # past value
        ;;
        --apiserver-version)
            checkInputParameterValue "${2}"
            APISERVER_VERSION="${2}"
            shift # past argument
            shift # past value
        ;;
        --oidc-host)
            checkInputParameterValue "${2}"
            OIDC_HOST="${2}"
            shift # past argument
            shift # past value
        ;;
        --oidc-client-id)
            checkInputParameterValue "${2}"
            OIDC_CLIENT_ID="${2}"
            shift # past argument
            shift # past value
        ;;
        --oidc-admin-group)
            checkInputParameterValue "${2}"
            OIDC_ADMIN_GROUP="${2}"
            shift # past argument
            shift # past value
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

function revert_migrator_file() {
    echo "$MIGRATOR_FILE" > "$ROOT_PATH"/chart/compass/templates/migrator-job.yaml
    echo "$UPDATE_EXPECTED_SCHEMA_VERSION_FILE" > "$ROOT_PATH"/chart/compass/templates/update-expected-schema-version-job.yaml
}

function set_oidc_config() {
  yq -i ".global.cockpit.auth.idpHost = \"$1\"" "$PATH_TO_VALUES"
  yq -i ".global.cockpit.auth.clientID = \"$2\"" "$PATH_TO_VALUES"
  if [[ -n ${3}  ]]; then
   yq -i ".adminGroupNames = \"$3\"" "$PATH_TO_HYDRATOR_VALUES"
  fi
}

# NOTE: Only one trap per script is supported.
function cleanup_trap() {
  if [[ -f k3d-ca.crt ]]; then
    rm -f k3d-ca.crt
  fi
  if [[ -f ${COMPASS_CERT_PATH} ]]; then
    rm -f "${COMPASS_CERT_PATH}"
  fi
  if [[ ${DUMP_DB} ]]; then
      revert_migrator_file
  fi
  if [[ ${RESET_VALUES_YAML} ]] ; then
    set_oidc_config "" "" "$DEFAULT_OIDC_ADMIN_GROUPS"
  fi

  pkill -P $$ || true # This MUST be at the end of the cleanup_trap function.
}

trap cleanup_trap RETURN EXIT INT TERM

function mount_k3d_ca_to_oathkeeper() {
  echo "Mounting k3d CA cert into oathkeeper's container..."

  docker exec k3d-kyma-server-0 cat /var/lib/rancher/k3s/server/tls/server-ca.crt > k3d-ca.crt
  kubectl create configmap -n kyma-system k3d-ca --from-file k3d-ca.crt --dry-run=client -o yaml | kubectl apply -f -

  OATHKEEPER_DEPLOYMENT_NAME=$(kubectl get deployment -n kyma-system | grep oathkeeper | awk '{print $1}')
  OATHKEEPER_CONTAINER_NAME=$(kubectl get deployment -n kyma-system "$OATHKEEPER_DEPLOYMENT_NAME" -o=jsonpath='{.spec.template.spec.containers[*].name}' | tr -s '[[:space:]]' '\n' | grep -v 'maester')

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"volumes":[{"configMap":{"defaultMode": 420,"name": "k3d-ca"},"name": "k3d-ca-volume"}]}}}}'

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"containers":[{"name": "'$OATHKEEPER_CONTAINER_NAME'","volumeMounts": [{ "mountPath": "'/etc/ssl/certs/k3d-ca.crt'","name": "k3d-ca-volume","subPath": "k3d-ca.crt"}]}]}}}}'
}

if [[ -z ${OIDC_HOST} || -z ${OIDC_CLIENT_ID} ]]; then
  if [[ -f ${PATH_TO_COMPASS_OIDC_CONFIG_FILE} ]]; then
    echo -e "${YELLOW}OIDC configuration not provided. Configuration from default config file will be used.${NC}"
    DEFAULT_OIDC_ADMIN_GROUPS="$(yq ".adminGroupNames" "$PATH_TO_HYDRATOR_VALUES")"
    OIDC_HOST=$(yq ".idpHost" "$PATH_TO_COMPASS_OIDC_CONFIG_FILE")
    OIDC_CLIENT_ID=$(yq ".clientID" "$PATH_TO_COMPASS_OIDC_CONFIG_FILE")
    OIDC_GROUPS=$(yq ".adminGroupNames" "$PATH_TO_COMPASS_OIDC_CONFIG_FILE")
    set_oidc_config "$OIDC_HOST" "$OIDC_CLIENT_ID" "$OIDC_GROUPS"
  else
    echo -e "${RED}OIDC configuration not provided and config file was not found. JWT flows will not work!${NC}"
    RESET_VALUES_YAML=false
  fi
else
  DEFAULT_OIDC_ADMIN_GROUPS="$(yq ".adminGroupNames" "$PATH_TO_HYDRATOR_VALUES")"
  if [[ -z ${OIDC_ADMIN_GROUP} ]]; then
    echo -e "${GREEN}Using provided OIDC host and client-id.${NC}"
    echo -e "${YELLOW}OIDC admin group was not provided. Will use default values.${NC}"
    set_oidc_config "$OIDC_HOST" "$OIDC_CLIENT_ID"
  else
    echo -e "${GREEN}Using provided OIDC host, client-id and admin group.${NC}"
    OIDC_GROUPS="$DEFAULT_OIDC_ADMIN_GROUPS , $OIDC_ADMIN_GROUP"
    set_oidc_config "$OIDC_HOST" "$OIDC_CLIENT_ID" "$OIDC_GROUPS"
  fi
fi

if [ -z "$KYMA_RELEASE" ]; then
  KYMA_RELEASE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)
fi

if [ -z "$KYMA_INSTALLATION" ]; then
  KYMA_INSTALLATION="minimal"
fi

if [[ ${DUMP_DB} ]]; then
    echo -e "${GREEN}DB dump will be used to prepopulate installation${NC}"

    if [ "$(uname)" == "Darwin" ]; then #  this is the case when the script is ran on local Mac OSX machines, reference issue: https://stackoverflow.com/questions/4247068/sed-command-with-i-option-failing-on-mac-but-works-on-linux
        sed -i '' 's/image\:.*compass-schema-migrator.*/image\: k3d-kyma-registry\:5001\/compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/migrator-job.yaml
        sed -i '' 's/image\:.*compass-schema-migrator.*/image\: k3d-kyma-registry\:5001\/compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/update-expected-schema-version-job.yaml
    else # this is the case when the script is ran on non-Mac OSX machines, ex. as part of remote PR jobs
        sed -i 's/image\:.*compass-schema-migrator.*/image\: k3d-kyma-registry\:5001\/compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/migrator-job.yaml
        sed -i 's/image\:.*compass-schema-migrator.*/image\: k3d-kyma-registry\:5001\/compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/update-expected-schema-version-job.yaml
    fi

    REMOTE_VERSIONS=($(gsutil ls -R gs://sap-cp-cmp-dev-db-dump/ | grep -o -E '[0-9]+' | sed -e 's/^0\+//' | sort -r))
    LOCAL_VERSIONS=($(ls migrations/director | grep -o -E '^[0-9]+' | sed -e 's/^0\+//' | sort -ru))

    SCHEMA_VERSION=""
    for r in "${REMOTE_VERSIONS[@]}"; do
       for l in "${LOCAL_VERSIONS[@]}"; do
          if [[ "$r" == "$l" ]]; then
            SCHEMA_VERSION=$r
            break 2;
          fi
       done
    done

# todo:: remove
#    LATEST_LOCAL_SCHEMA_VERSION=$(ls migrations/director | tail -n 1 | grep -o -E '^[0-9]+' | sed -e 's/^0\+//')
#    echo -e "${YELLOW}Check if there is DB dump in GCS bucket with migration number: $LATEST_LOCAL_SCHEMA_VERSION...${NC}"
#    gsutil -q stat gs://sap-cp-cmp-dev-db-dump/dump-"${LATEST_LOCAL_SCHEMA_VERSION}".sql
#    STATUS=$?
#
#    if [[ ! $STATUS ]]; then
#      echo -e "${YELLOW}There is no DB dump with migration number: $LATEST_LOCAL_SCHEMA_VERSION in the bucket. Will get the latest available one...${NC}"
#      LATEST_DB_DUMP_VERSION=$(gsutil ls -R gs://sap-cp-cmp-dev-db-dump/ | tail -n 1 | grep -o -E '[0-9]+' | sed -e 's/^0\+//')
#      SCHEMA_VERSION=$LATEST_DB_DUMP_VERSION
#    else
#      echo -e "${GREEN}DB dump with migration number: $LATEST_LOCAL_SCHEMA_VERSION exists in the bucket. Will use it...${NC}"
#      SCHEMA_VERSION=$LATEST_LOCAL_SCHEMA_VERSION
#    fi

    echo -e "${YELLOW}Check if there is DB dump in GCS bucket with migration number: $SCHEMA_VERSION...${NC}"
    gsutil -q stat gs://sap-cp-cmp-dev-db-dump/dump-"${SCHEMA_VERSION}".sql
    STATUS=$?

    if [[ $STATUS ]]; then
      echo -e "${GREEN}DB dump with migration number: $SCHEMA_VERSION exists in the bucket. Will use it...${NC}"
    else
      echo -e "${RED}There is no DB dump with migration number: $SCHEMA_VERSION in the bucket.${NC}"
      exit 1
    fi

    if [[ ! -f ${SCHEMA_MIGRATOR_COMPONENT_PATH}/seeds/dump-${SCHEMA_VERSION}.sql ]]; then
        echo -e "${YELLOW}There is no dump with number: $SCHEMA_VERSION locally. Will pull the DB dump from GCR bucket...${NC}"
        gsutil cp gs://sap-cp-cmp-dev-db-dump/dump-"${SCHEMA_VERSION}".sql "$SCHEMA_MIGRATOR_COMPONENT_PATH"/seeds/dump-"${SCHEMA_VERSION}".sql
    else
        echo -e "${GREEN}DB dump already exists on the local system, will reuse it${NC}"
    fi
fi

if [[ ! ${SKIP_K3D_START} ]]; then
  echo "Provisioning k3d cluster..."
  # todo cpu limit
  kyma provision k3d \
  --k3s-arg '--kube-apiserver-arg=anonymous-auth=true@server:*' \
  --k3s-arg '--kube-apiserver-arg=feature-gates=ServiceAccountIssuerDiscovery=true@server:*' \
  --k3d-arg='--servers-memory '"${K3D_MEMORY}" \
  --k3d-arg='--agents-memory '"${K3D_MEMORY}" \
  --timeout "${K3D_TIMEOUT}" \
  --kube-version "${APISERVER_VERSION}"
  echo "Adding k3d registry entry to /etc/hosts..."
  sudo sh -c "echo \"\n127.0.0.1 k3d-kyma-registry\" >> /etc/hosts"
fi

usek3d

echo "Label k3d node for benchmark execution..."
NODE=$(kubectl get nodes | grep agent | tail -n 1 | cut -d ' ' -f 1)
kubectl label --overwrite node "$NODE" benchmark=true || true

if [[ ${DUMP_DB} ]]; then
    echo -e "${YELLOW}DUMP_DB option is selected. Building an image for the schema-migrator using local files...${NC}"
    export DOCKER_TAG=$DUMP_IMAGE_TAG
    make -C ${ROOT_PATH}/components/schema-migrator build-for-k3d
fi

if [[ ! ${SKIP_KYMA_START} ]]; then
  LOCAL_ENV=true bash "${ROOT_PATH}"/installation/scripts/install-kyma.sh --kyma-release ${KYMA_RELEASE} --kyma-installation ${KYMA_INSTALLATION}
  kubectl set image -n kyma-system cronjob/oathkeeper-jwks-rotator keys-generator=oryd/oathkeeper:v0.38.23
  kubectl patch cronjob -n kyma-system oathkeeper-jwks-rotator -p '{"spec":{"schedule": "*/1 * * * *"}}'
  until [[ $(kubectl get cronjob -n kyma-system oathkeeper-jwks-rotator --output=jsonpath={.status.lastScheduleTime}) ]]; do
      echo "Waiting for cronjob oathkeeper-jwks-rotator to be scheduled"
      sleep 3
  done
  kubectl patch cronjob -n kyma-system oathkeeper-jwks-rotator -p '{"spec":{"schedule": "0 0 1 * *"}}'
fi

mount_k3d_ca_to_oathkeeper

# Currently there is a problem fetching JWKS keys, used to validate JWT token send to hydra. The function bellow patches the RequestAuthentication istio resource
# with the needed keys, by first getting them using kubectl
function patchJWKS() {
  JWKS="'$(kubectl get --raw '/openid/v1/jwks')'"
  until [[ $(kubectl get requestauthentication kyma-internal-authn -n kyma-system 2>/dev/null) &&
          $(kubectl get requestauthentication compass-internal-authn -n compass-system 2>/dev/null) ]]; do
    echo "Waiting for requestauthentication resources to be created"
    sleep 3
  done
  kubectl get requestauthentication kyma-internal-authn -n kyma-system -o yaml | sed 's/jwksUri\:.*$/jwks\: '$JWKS'/' | kubectl apply -f -
  kubectl get requestauthentication compass-internal-authn -n compass-system -o yaml | sed 's/jwksUri\:.*$/jwks\: '$JWKS'/' | kubectl apply -f -
}
patchJWKS&

echo 'Installing Compass'
COMPASS_OVERRIDES="${CURRENT_DIR}/../resources/compass-overrides-local.yaml"
bash "${ROOT_PATH}"/installation/scripts/install-compass.sh --overrides-file "${COMPASS_OVERRIDES}" --timeout 30m0s
STATUS=$(helm status compass -n compass-system -o json | jq .info.status)
echo "Compass installation status ${STATUS}"

prometheusMTLSPatch

echo 'Adding compass certificate to keychain'
COMPASS_CERT_PATH="${CURRENT_DIR}/../cmd/compass-cert.pem"
echo -n | openssl s_client -showcerts -servername compass.local.kyma.dev -connect compass.local.kyma.dev:443 2>/dev/null | openssl x509 -inform pem > "${COMPASS_CERT_PATH}"
if [ "$(uname)" == "Darwin" ]; then #  this is the case when the script is ran on local Mac OSX machines
  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "${COMPASS_CERT_PATH}"
else # this is the case when the script is ran on non-Mac OSX machines, ex. as part of remote PR jobs
  sudo cp "${COMPASS_CERT_PATH}" /etc/ssl/certs
  sudo update-ca-certificates
fi

echo "Adding Compass entries to /etc/hosts..."
sudo sh -c "echo \"\n127.0.0.1 adapter-gateway.local.kyma.dev adapter-gateway-mtls.local.kyma.dev compass-gateway-mtls.local.kyma.dev compass-gateway-xsuaa.local.kyma.dev compass-gateway-sap-mtls.local.kyma.dev compass-gateway-auth-oauth.local.kyma.dev compass-gateway.local.kyma.dev compass-gateway-int.local.kyma.dev compass.local.kyma.dev compass-mf.local.kyma.dev kyma-env-broker.local.kyma.dev director.local.kyma.dev compass-external-services-mock.local.kyma.dev compass-external-services-mock-sap-mtls.local.kyma.dev compass-external-services-mock-sap-mtls-ord.local.kyma.dev compass-external-services-mock-sap-mtls-global-ord-registry.local.kyma.dev\" >> /etc/hosts"