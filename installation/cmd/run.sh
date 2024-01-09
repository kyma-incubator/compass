#!/usr/bin/env bash

set -o errexit

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
DATA_DIR="${CURRENT_DIR}/../data"
mkdir -p "${DATA_DIR}"
source $SCRIPTS_DIR/utils.sh
source $SCRIPTS_DIR/prom-mtls-patch.sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..
PATH_TO_VALUES="$CURRENT_DIR/../../chart/compass/values.yaml"
PATH_TO_HYDRATOR_VALUES="$CURRENT_DIR/../../chart/compass/charts/hydrator/values.yaml"
PATH_TO_COMPASS_OIDC_CONFIG_FILE="$HOME/.compass.yaml"
UPDATE_EXPECTED_SCHEMA_VERSION_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/update-expected-schema-version-job.yaml)
SCHEMA_MIGRATOR_COMPONENT_PATH=${ROOT_PATH}/components/schema-migrator
RESET_VALUES_YAML=true

K3D_NAME="kyma"
K3D_MEMORY=8192MB
K3D_TIMEOUT=10m0s
APISERVER_VERSION=1.25.6

# These variables are used only during local installation to override the utils to use the k3d cluster
KUBECTL="kubectl_k3d_kyma"
HELM="helm_k3d_kyma"
KYMA="kyma_k3d_kyma"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --skip-k3d-start)
            SKIP_K3D_START=true
            shift # past argument
        ;;
        --skip-kyma-start)
            SKIP_KYMA_START=true
            shift # past argument
        ;;
        --skip-db-install)
            SKIP_DB_INSTALL=true
            shift # past argument
        ;;
        --skip-ory-install)
            SKIP_ORY_INSTALL=true
            shift # past argument
        ;;
        --dump-db)
            DUMP_DB=true
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
  if [[ -f "$K3D_CA" ]]; then
    rm -f "$K3D_CA"
  fi
  if [[ -f ${COMPASS_CERT_PATH} ]]; then
    rm -f "${COMPASS_CERT_PATH}"
  fi
  if [[ ${DUMP_DB} ]]; then
      revert_migrator_file
      rm -rf "${DATA_DIR}"/dump || true
  fi
  if [[ ${RESET_VALUES_YAML} ]] ; then
    set_oidc_config "" "" "$DEFAULT_OIDC_ADMIN_GROUPS"
  fi
  pkill -P $$ || true # This MUST be at the end of the cleanup_trap function.
}

trap cleanup_trap RETURN EXIT INT TERM

K3D_CA=k3d-ca.crt
function mount_k3d_ca_to_oathkeeper() {
  echo "Mounting k3d CA cert into oathkeeper's container..."

  docker exec k3d-kyma-server-0 cat /var/lib/rancher/k3s/server/tls/server-ca.crt > "$K3D_CA"
  "$KUBECTL" create configmap -n ory k3d-ca --from-file "$K3D_CA" --dry-run=client -o yaml | "$KUBECTL" apply -f -

  OATHKEEPER_DEPLOYMENT_NAME=$("$KUBECTL" get deployment -n ory | grep oathkeeper | awk '{print $1}')
  OATHKEEPER_CONTAINER_NAME=$("$KUBECTL" get deployment -n ory "$OATHKEEPER_DEPLOYMENT_NAME" -o=jsonpath='{.spec.template.spec.containers[*].name}' | tr -s '[[:space:]]' '\n' | grep -v 'maester')

  "$KUBECTL" -n ory patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
   -p '{"spec":{"template":{"spec":{"containers":[{"name": "'$OATHKEEPER_CONTAINER_NAME'","volumeMounts": [{ "mountPath": "'/etc/ssl/certs/k3d-ca.crt'","name": "k3d-ca-volume","subPath": "k3d-ca.crt"}]}],"volumes":[{"configMap":{"defaultMode": 420,"name": "k3d-ca"},"name": "k3d-ca-volume"}]}}}}'
}

# Currently there is a problem fetching JWKS keys, used to validate JWT token send to hydra. The function bellow patches the RequestAuthentication istio resource
# with the needed keys, by first getting them using kubectl
function patchJWKS() {
  echo "Patching Request Authentication resources..."
  JWKS="'$("$KUBECTL" get --raw '/openid/v1/jwks')'"
  until [[ $("$KUBECTL" get requestauthentication ory-internal-authn -n ory 2>/dev/null) &&
          $("$KUBECTL" get requestauthentication compass-internal-authn -n compass-system 2>/dev/null) ]]; do
    echo "Waiting for Request Authentication resources to be created"
    sleep 8
  done
  "$KUBECTL" get requestauthentication ory-internal-authn -n ory -o yaml | sed 's/jwksUri\:.*$/jwks\: '$JWKS'/' | "$KUBECTL" apply -f -
  "$KUBECTL" get requestauthentication compass-internal-authn -n compass-system -o yaml | sed 's/jwksUri\:.*$/jwks\: '$JWKS'/' | "$KUBECTL" apply -f -
  echo "Request Authentication resources were successfully patched"
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

if [[ ${DUMP_DB} ]]; then
    echo -e "${GREEN}DB dump will be used to prepopulate installation${NC}"

    REMOTE_VERSIONS=($(gsutil ls -R gs://sap-cp-cmp-dev-db-dump/ | grep -o -E '[0-9]+' | sed -e 's/^0\+//' | sort -r))
    LOCAL_VERSIONS=($(ls "$SCHEMA_MIGRATOR_COMPONENT_PATH"/migrations/director | grep -o -E '^[0-9]+' | sed -e 's/^0\+//' | sort -ru))

    SCHEMA_VERSION=""
    for r in "${REMOTE_VERSIONS[@]}"; do
       for l in "${LOCAL_VERSIONS[@]}"; do
          if [[ "$r" == "$l" ]]; then
            SCHEMA_VERSION=$r
            break 2;
          fi
       done
    done

    if [[ -z $SCHEMA_VERSION ]]; then
      echo -e "${RED}\$SCHEMA_VERSION variable cannot be empty${NC}"
    fi

    echo -e "${YELLOW}Check if there is DB dump in GCS bucket with migration number: $SCHEMA_VERSION...${NC}"
    gsutil -q stat gs://sap-cp-cmp-dev-db-dump/dump-"${SCHEMA_VERSION}"/toc.dat
    STATUS=$?

    if [[ $STATUS ]]; then
      echo -e "${GREEN}DB dump with migration number: $SCHEMA_VERSION exists in the bucket. Will use it...${NC}"
    else
      echo -e "${RED}There is no DB dump with migration number: $SCHEMA_VERSION in the bucket.${NC}"
      exit 1
    fi

    if [[ ! -d ${DATA_DIR}/dump-${SCHEMA_VERSION} ]]; then
        echo -e "${YELLOW}There is no dump with number: $SCHEMA_VERSION locally. Will pull the DB dump from GCR bucket...${NC}"
        mkdir ${DATA_DIR}/dump-${SCHEMA_VERSION}
        gsutil cp -r gs://sap-cp-cmp-dev-db-dump/dump-"${SCHEMA_VERSION}" "${DATA_DIR}"
    else
        echo -e "${GREEN}DB dump already exists on the local system, will reuse it${NC}"
    fi
    rm -rf "${DATA_DIR}"/dump || true
    cp -R "${DATA_DIR}"/dump-"${SCHEMA_VERSION}" "${DATA_DIR}"/dump
fi

# If the script is not run by the root user then 'sudo' should be used for some commands
# This is necessary as this script is ran for local installation and by PR jobs
# The PR jobs use alpine that does not have 'sudo' as a command and always executes scripts as the root user
SUDO=''
if (( $EUID != 0 )); then
    SUDO='sudo'
fi

if [[ ! ${SKIP_K3D_START} ]]; then
  echo "Provisioning k3d cluster..."
  kyma provision k3d \
  --name "$K3D_NAME" \
  --k3s-arg '--kube-apiserver-arg=anonymous-auth=true@server:*' \
  --k3d-arg='--servers-memory '"${K3D_MEMORY}" \
  --k3d-arg='--agents-memory '"${K3D_MEMORY}" \
  --k3d-arg='--volume '"${DATA_DIR}:/tmp/dbdata" \
  --timeout "${K3D_TIMEOUT}" \
  --kube-version "${APISERVER_VERSION}"
  echo "Adding k3d registry entry to /etc/hosts..."
  $SUDO sh -c "echo \"\n127.0.0.1 k3d-kyma-registry\" >> /etc/hosts"
fi

echo "Label k3d node for benchmark execution..."
NODE=$("$KUBECTL" get nodes | grep agent | tail -n 1 | cut -d ' ' -f 1)
"$KUBECTL" label --overwrite node "$NODE" benchmark=true || true

if [[ ! ${SKIP_KYMA_START} ]]; then
  KYMA="$KYMA" KUBECTL="$KUBECTL" LOCAL_ENV=true bash "${ROOT_PATH}"/installation/scripts/install-kyma.sh
fi

if [[ ! ${SKIP_DB_INSTALL} ]]; then
  DB_OVERRIDES="${CURRENT_DIR}/../resources/compass-overrides-local.yaml"
  KUBECTL="$KUBECTL" HELM="$HELM" bash "${ROOT_PATH}"/installation/scripts/install-db.sh --overrides-file "${DB_OVERRIDES}" --timeout 30m0s
  STATUS=$("$HELM" status localdb -n compass-system -o json | jq .info.status)
  echo "DB installation status ${STATUS}"
fi

if [[ ! ${SKIP_ORY_INSTALL} ]]; then
  echo "Installing ORY Stack..."
  KUBECTL="$KUBECTL" HELM="$HELM" bash "${ROOT_PATH}"/installation/scripts/install-ory.sh
fi

if [[ ! "$("$HELM" status ory-stack -n ory)" ]]; then
  echo -e "${RED}Ory Helm release does not exist, please omit the '--skip-ory-install' to install it.${NC}"
  exit 1
fi

mount_k3d_ca_to_oathkeeper

patchJWKS&

COMPASS_OVERRIDES="${CURRENT_DIR}/../resources/compass-overrides-local.yaml"
KUBECTL="$KUBECTL" HELM="$HELM" bash "${ROOT_PATH}"/installation/scripts/install-compass.sh --overrides-file "${COMPASS_OVERRIDES}" --timeout 30m0s --sql-helm-backend

prometheusMTLSPatch

echo 'Adding compass certificate to keychain'
COMPASS_CERT_PATH="${CURRENT_DIR}/../cmd/compass-cert.pem"
echo -n | openssl s_client -showcerts -servername compass.local.kyma.dev -connect compass.local.kyma.dev:443 2>/dev/null | openssl x509 -inform pem > "${COMPASS_CERT_PATH}"
if [ "$(uname)" == "Darwin" ]; then #  this is the case when the script is ran on local Mac OSX machines
  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "${COMPASS_CERT_PATH}"
else # this is the case when the script is ran on non-Mac OSX machines, ex. as part of remote PR jobs
  $SUDO cp "${COMPASS_CERT_PATH}" /etc/ssl/certs
  $SUDO update-ca-certificates
fi

echo "Adding Compass entries to /etc/hosts..."
$SUDO sh -c "echo \"\n127.0.0.1 adapter-gateway.local.kyma.dev adapter-gateway-mtls.local.kyma.dev compass-gateway-mtls.local.kyma.dev compass-gateway-xsuaa.local.kyma.dev compass-gateway-sap-mtls.local.kyma.dev compass-gateway-auth-oauth.local.kyma.dev compass-gateway.local.kyma.dev compass-gateway-int.local.kyma.dev compass.local.kyma.dev compass-mf.local.kyma.dev kyma-env-broker.local.kyma.dev director.local.kyma.dev compass-external-services-mock.local.kyma.dev compass-external-services-mock-sap-mtls.local.kyma.dev compass-external-services-mock-sap-mtls-ord.local.kyma.dev compass-external-services-mock-sap-mtls-global-ord-registry.local.kyma.dev discovery.api.local compass-director-internal.local.kyma.dev connector.local.kyma.dev hydrator.local.kyma.dev compass-gateway-internal.local.kyma.dev\" >> /etc/hosts"
