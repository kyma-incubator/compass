#!/usr/bin/env bash

set -o errexit

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh
source $SCRIPTS_DIR/prom-mtls-patch.sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..
MIGRATOR_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/migrator-job.yaml)
UPDATE_EXPECTED_SCHEMA_VERSION_FILE=$(cat "$ROOT_PATH"/chart/compass/templates/update-expected-schema-version-job.yaml)

K3D_MEMORY=8192MB
K3D_TIMEOUT=10m0s
K3D_CPUS=5
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
        --k3d-cpus)
            checkInputParameterValue "${2}"
            K3D_CPUS="${2}"
            shift # past argument
            shift # past value
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

function mount_k3d_ca_to_oathkeeper() {
  echo "Mounting k3d CA cert into oathkeeper's container..."

  docker exec k3d-kyma-server-0 cat /var/lib/rancher/k3s/server/tls/server-ca.crt > k3d-ca.crt
  trap "rm -f k3d-ca.crt" RETURN EXIT INT TERM

  kubectl create configmap -n kyma-system k3d-ca --from-file k3d-ca.crt --dry-run=client -o yaml | kubectl apply -f -

  OATHKEEPER_DEPLOYMENT_NAME=$(kubectl get deployment -n kyma-system | grep oathkeeper | awk '{print $1}')
  OATHKEEPER_CONTAINER_NAME=$(kubectl get deployment -n kyma-system "$OATHKEEPER_DEPLOYMENT_NAME" -o=jsonpath='{.spec.template.spec.containers[*].name}' | tr -s '[[:space:]]' '\n' | grep -v 'maester')

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"volumes":[{"configMap":{"defaultMode": 420,"name": "k3d-ca"},"name": "k3d-ca-volume"}]}}}}'

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"containers":[{"name": "'$OATHKEEPER_CONTAINER_NAME'","volumeMounts": [{ "mountPath": "'/etc/ssl/certs/k3d-ca.crt'","name": "k3d-ca-volume","subPath": "k3d-ca.crt"}]}]}}}}'
}

if [[ ${DUMP_DB} ]]; then
    trap revert_migrator_file EXIT
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


echo "Label k3d node for benchmark execution..."
NODE=$(kubectl get nodes | grep agent | tail -n 1 | cut -d ' ' -f 1)
kubectl label --overwrite node "$NODE" benchmark=true || true

if [[ ${DUMP_DB} ]]; then
    echo -e "${YELLOW}DUMP_DB option is selected. Building an image for the schema-migrator using local files...${NC}"
    export DOCKER_TAG=$DUMP_IMAGE_TAG
    make -C ${ROOT_PATH}/components/schema-migrator build-to-minikube
fi

if [[ ! ${SKIP_KYMA_START} ]]; then
  LOCAL_ENV=true bash "${ROOT_PATH}"/installation/scripts/install-kyma.sh --kyma-release ${KYMA_RELEASE} --kyma-installation ${KYMA_INSTALLATION}
fi

#if [[ `kubectl get TestDefinition logging -n kyma-system` ]]; then
#  # Patch logging TestDefinition
#  HOSTNAMES_TO=$(kubectl get TestDefinition logging -n kyma-system -o json |  jq -r '.spec.template.spec.hostAliases[0].hostnames += ["loki.local.kyma.dev"]' | jq -r ".spec.template.spec.hostAliases[0].hostnames" | jq ". | unique")
#  IP_TO=$(kubectl get TestDefinition logging -n kyma-system -o json | jq '.spec.template.spec.hostAliases[0].ip')
#  PATCH_TO=$(cat "${ROOT_PATH}"/installation/resources/logging-test-definition-patch.json | jq -c ".spec.template.spec.hostAliases[0].hostnames += $HOSTNAMES_TO" | jq -c ".spec.template.spec.hostAliases[0].ip += $IP_TO")
#  kubectl patch TestDefinition logging -n kyma-system  --type='merge' -p "$PATCH_TO"
#fi

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
bash "${ROOT_PATH}"/installation/scripts/install-compass.sh
STATUS=$(helm status compass -n compass-system -o json | jq .info.status)
echo "Compass installation status ${STATUS}"

prometheusMTLSPatch

echo "Adding Compass entries to /etc/hosts..."
sudo sh -c "echo \"\n127.0.0.1 adapter-gateway.local.kyma.dev adapter-gateway-mtls.local.kyma.dev compass-gateway-mtls.local.kyma.dev compass-gateway-sap-mtls.local.kyma.dev compass-gateway-auth-oauth.local.kyma.dev compass-gateway.local.kyma.dev compass-gateway-int.local.kyma.dev compass.local.kyma.dev compass-mf.local.kyma.dev kyma-env-broker.local.kyma.dev director.local.kyma.dev compass-external-services-mock-sap-mtls.local.kyma.dev\" >> /etc/hosts"