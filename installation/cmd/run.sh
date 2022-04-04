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
PATH_TO_DIRECTOR_VALUES="$CURRENT_DIR/../../chart/compass/charts/director/values.yaml"
PATH_TO_COMPASS_OIDC_CONFIG_FILE="$HOME/.compass.yaml"
MIGRATOR_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/migrator-job.yaml)
UPDATE_EXPECTED_SCHEMA_VERSION_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/update-expected-schema-version-job.yaml)
RESET_VALUES_YAML=true

MINIKUBE_MEMORY=8192
MINIKUBE_TIMEOUT=25m
MINIKUBE_CPUS=5
APISERVER_VERSION=1.19.16


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
        --skip-minikube-start)
            SKIP_MINIKUBE_START=true
            shift # past argument
        ;;
        --skip-kyma-start)
            SKIP_KYMA_START=true
            shift # past argument
        ;;
        --docker-driver)
            DOCKER_DRIVER=true
            shift # past argument
        ;;
        --dump-db)
            DUMP_DB=true
            DUMP_IMAGE_TAG="dump"
            shift # past argument
        ;;
        --minikube-cpus)
            checkInputParameterValue "${2}"
            MINIKUBE_CPUS="${2}"
            shift # past argument
            shift # past value
        ;;
        --minikube-memory)
            checkInputParameterValue "${2}"
            MINIKUBE_MEMORY="${2}"
            shift # past argument
            shift # past value
        ;;
        --minikube-timeout)
            checkInputParameterValue "${2}"
            MINIKUBE_TIMEOUT="${2}"
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
   yq -i ".adminGroupNames = \"$3\"" "$PATH_TO_DIRECTOR_VALUES"
  fi
}

function cleanup_trap() {
  if [[ -f mk-ca.crt ]]; then
    rm -f mk-ca.crt
  fi
  if [[ ${DUMP_DB} ]]; then
      revert_migrator_file
  fi
  if $RESET_VALUES_YAML ; then
    set_oidc_config "" "" "$DEFAULT_OIDC_ADMIN_GROUPS"
  fi
}

function mount_minikube_ca_to_oathkeeper() {
  echo "Mounting minikube CA cert into oathkeeper's container..."

  cat $HOME/.minikube/ca.crt > mk-ca.crt

  kubectl create configmap -n kyma-system minikube-ca --from-file mk-ca.crt --dry-run -o yaml | kubectl apply -f -

  OATHKEEPER_DEPLOYMENT_NAME=$(kubectl get deployment -n kyma-system | grep oathkeeper | awk '{print $1}')
  OATHKEEPER_CONTAINER_NAME=$(kubectl get deployment -n kyma-system "$OATHKEEPER_DEPLOYMENT_NAME" -o=jsonpath='{.spec.template.spec.containers[*].name}' | tr -s '[[:space:]]' '\n' | grep -v 'maester')

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"volumes":[{"configMap":{"defaultMode": 420,"name": "minikube-ca"},"name": "minikube-ca-volume"}]}}}}'

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"containers":[{"name": "'$OATHKEEPER_CONTAINER_NAME'","volumeMounts": [{ "mountPath": "'/etc/ssl/certs/mk-ca.crt'","name": "minikube-ca-volume","subPath": "mk-ca.crt"}]}]}}}}'
}

if [[ -z ${OIDC_HOST} || -z ${OIDC_CLIENT_ID} ]]; then
  if [[ -f ${PATH_TO_COMPASS_OIDC_CONFIG_FILE} ]]; then
    echo -e "${YELLOW}OIDC configuration not provided. Configuration from default config file will be used.${NC}"
    DEFAULT_OIDC_ADMIN_GROUPS="$(yq ".adminGroupNames" "$PATH_TO_DIRECTOR_VALUES")"
    OIDC_HOST=$(yq ".idpHost" "$PATH_TO_COMPASS_OIDC_CONFIG_FILE")
    OIDC_CLIENT_ID=$(yq ".clientID" "$PATH_TO_COMPASS_OIDC_CONFIG_FILE")
    OIDC_GROUPS=$(yq ".adminGroupNames" "$PATH_TO_COMPASS_OIDC_CONFIG_FILE")
    set_oidc_config "$OIDC_HOST" "$OIDC_CLIENT_ID" "$OIDC_GROUPS"
  else
    echo -e "${RED}OIDC configuration not provided and config file was not found. JWT flows will not work!${NC}"
    RESET_VALUES_YAML=false
  fi
else
  DEFAULT_OIDC_ADMIN_GROUPS="$(yq ".adminGroupNames" "$PATH_TO_DIRECTOR_VALUES")"
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

trap "cleanup_trap" RETURN EXIT INT TERM

if [ -z "$KYMA_RELEASE" ]; then
  KYMA_RELEASE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)
fi

if [ -z "$KYMA_INSTALLATION" ]; then
  KYMA_INSTALLATION="minimal"
fi

if [[ ${DUMP_DB} ]]; then
    echo -e "${GREEN}DB dump will be used to prepopulate installation${NC}"

    if [ "$(uname)" == "Darwin" ]; then #  this is the case when the script is ran on local Mac OSX machines, reference issue: https://stackoverflow.com/questions/4247068/sed-command-with-i-option-failing-on-mac-but-works-on-linux
        sed -i '' 's/image\:.*compass-schema-migrator.*/image\: compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/migrator-job.yaml
        sed -i '' 's/image\:.*compass-schema-migrator.*/image\: compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/update-expected-schema-version-job.yaml
    else # this is the case when the script is ran on non-Mac OSX machines, ex. as part of remote PR jobs
        sed -i 's/image\:.*compass-schema-migrator.*/image\: compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/migrator-job.yaml
        sed -i 's/image\:.*compass-schema-migrator.*/image\: compass-schema-migrator\:'$DUMP_IMAGE_TAG'/' ${ROOT_PATH}/chart/compass/templates/update-expected-schema-version-job.yaml
    fi


    if [[ ! -f ${ROOT_PATH}/components/schema-migrator/seeds/dump.sql ]]; then
        echo -e "${YELLOW}Will pull DB dump from GCR bucket${NC}"
        gsutil cp gs://sap-cp-cmp-dev-db-dump/dump.sql ${ROOT_PATH}/components/schema-migrator/seeds/dump.sql
    else
        echo -e "${GREEN}DB dump already exists on system, will reuse it${NC}"
    fi
fi

if [[ ! ${SKIP_MINIKUBE_START} ]]; then
  echo "Provisioning Minikube cluster..."
  if [[ ! ${DOCKER_DRIVER} ]]; then
    kyma provision minikube --cpus ${MINIKUBE_CPUS} --memory ${MINIKUBE_MEMORY} --timeout ${MINIKUBE_TIMEOUT} --kube-version ${APISERVER_VERSION}
  else
    kyma provision minikube --cpus ${MINIKUBE_CPUS} --memory ${MINIKUBE_MEMORY} --timeout ${MINIKUBE_TIMEOUT} --kube-version ${APISERVER_VERSION} --vm-driver docker --docker-ports 443:443 --docker-ports 80:80
  fi
fi

useMinikube

echo "Label Minikube node for benchmark execution..."
NODE=$(kubectl get nodes | tail -n 1 | cut -d ' ' -f 1)
kubectl label node "$NODE" benchmark=true || true

if [[ ${DUMP_DB} ]]; then
    echo -e "${YELLOW}DUMP_DB option is selected. Building an image for the schema-migrator using local files...${NC}"
    export DOCKER_TAG=$DUMP_IMAGE_TAG
    make -C ${ROOT_PATH}/components/schema-migrator build-to-minikube
fi

if [[ ! ${SKIP_KYMA_START} ]]; then
  LOCAL_ENV=true bash "${ROOT_PATH}"/installation/scripts/install-kyma.sh --kyma-release ${KYMA_RELEASE} --kyma-installation ${KYMA_INSTALLATION}
fi

if [[ `kubectl get TestDefinition logging -n kyma-system` ]]; then
  # Patch logging TestDefinition
  HOSTNAMES_TO=$(kubectl get TestDefinition logging -n kyma-system -o json |  jq -r '.spec.template.spec.hostAliases[0].hostnames += ["loki.kyma.local"]' | jq -r ".spec.template.spec.hostAliases[0].hostnames" | jq ". | unique")
  IP_TO=$(kubectl get TestDefinition logging -n kyma-system -o json | jq '.spec.template.spec.hostAliases[0].ip')
  PATCH_TO=$(cat "${ROOT_PATH}"/installation/resources/logging-test-definition-patch.json | jq -c ".spec.template.spec.hostAliases[0].hostnames += $HOSTNAMES_TO" | jq -c ".spec.template.spec.hostAliases[0].ip += $IP_TO")
  kubectl patch TestDefinition logging -n kyma-system  --type='merge' -p "$PATCH_TO"
fi

if [[ `kubectl get TestDefinition dex-connection -n kyma-system` ]]; then
  # Patch dex-connection TestDefinition
  kubectl patch TestDefinition dex-connection -n kyma-system --type=json -p="[{\"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/image\", \"value\": \"eu.gcr.io/kyma-project/external/curlimages/curl:7.70.0\"}]"
fi

mount_minikube_ca_to_oathkeeper

prometheusMTLSPatch

bash "${ROOT_PATH}"/installation/scripts/run-compass-installer.sh --kyma-installation ${KYMA_INSTALLATION}
sleep 15
bash "${ROOT_PATH}"/installation/scripts/is-installed.sh

echo "Adding Compass entries to /etc/hosts..."

MINIKUBE_IP=$(minikube ip)
if [[ ${DOCKER_DRIVER} ]]; then
    MINIKUBE_IP=127.0.0.1
fi
sudo sh -c "echo \"\n${MINIKUBE_IP} adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-xsuaa.kyma.local compass-gateway-sap-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass-gateway-int.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local director.kyma.local compass-external-services-mock.kyma.local compass-external-services-mock-sap-mtls.kyma.local compass-external-services-mock-sap-mtls-ord.kyma.local compass-external-services-mock-sap-mtls-global-ord-registry.kyma.local\" >> /etc/hosts"
