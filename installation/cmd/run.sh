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

function mount_minikube_ca_to_oathkeeper() {
  echo "Mounting minikube CA cert into oathkeeper's container..."

  cat $HOME/.minikube/ca.crt > mk-ca.crt
  trap "rm -f mk-ca.crt" RETURN EXIT INT TERM

  kubectl create configmap -n kyma-system minikube-ca --from-file mk-ca.crt --dry-run -o yaml | kubectl apply -f -

  OATHKEEPER_DEPLOYMENT_NAME=$(kubectl get deployment -n kyma-system | grep oathkeeper | awk '{print $1}')
  OATHKEEPER_CONTAINER_NAME=$(kubectl get deployment -n kyma-system "$OATHKEEPER_DEPLOYMENT_NAME" -o=jsonpath='{.spec.template.spec.containers[*].name}' | tr -s '[[:space:]]' '\n' | grep -v 'maester')

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"volumes":[{"configMap":{"defaultMode": 420,"name": "minikube-ca"},"name": "minikube-ca-volume"}]}}}}'

  kubectl -n kyma-system patch deployment "$OATHKEEPER_DEPLOYMENT_NAME" \
 -p '{"spec":{"template":{"spec":{"containers":[{"name": "'$OATHKEEPER_CONTAINER_NAME'","volumeMounts": [{ "mountPath": "'/etc/ssl/certs/mk-ca.crt'","name": "minikube-ca-volume","subPath": "mk-ca.crt"}]}]}}}}'
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
  kyma provision k3d --k3d-arg='--servers-memory '${K3D_MEMORY} --k3d-arg='--agents-memory '${K3D_MEMORY} --timeout ${K3D_TIMEOUT} --kube-version "${APISERVER_VERSION}"
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

# todo
#if [[ `kubectl get TestDefinition logging -n kyma-system` ]]; then
#  # Patch logging TestDefinition
#  HOSTNAMES_TO=$(kubectl get TestDefinition logging -n kyma-system -o json |  jq -r '.spec.template.spec.hostAliases[0].hostnames += ["loki.kyma.local"]' | jq -r ".spec.template.spec.hostAliases[0].hostnames" | jq ". | unique")
#  IP_TO=$(kubectl get TestDefinition logging -n kyma-system -o json | jq '.spec.template.spec.hostAliases[0].ip')
#  PATCH_TO=$(cat "${ROOT_PATH}"/installation/resources/logging-test-definition-patch.json | jq -c ".spec.template.spec.hostAliases[0].hostnames += $HOSTNAMES_TO" | jq -c ".spec.template.spec.hostAliases[0].ip += $IP_TO")
#  kubectl patch TestDefinition logging -n kyma-system  --type='merge' -p "$PATCH_TO"
#fi
#
#if [[ `kubectl get TestDefinition dex-connection -n kyma-system` ]]; then
#  # Patch dex-connection TestDefinition
#  kubectl patch TestDefinition dex-connection -n kyma-system --type=json -p="[{\"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/image\", \"value\": \"eu.gcr.io/kyma-project/external/curlimages/curl:7.70.0\"}]"
#fi

#mount_minikube_ca_to_oathkeeper

#prometheusMTLSPatch

echo 'Installing Compass'
bash "${ROOT_PATH}"/installation/scripts/install-compass.sh
sleep 5
helm status compass -o json

echo "Adding Compass entries to /etc/hosts..."
K3D_IP=127.0.0.1
sudo sh -c "echo \"\n${K3D_IP} adapter-gateway.local.kyma.dev adapter-gateway-mtls.local.kyma.dev compass-gateway-mtls.local.kyma.dev compass-gateway-sap-mtls.local.kyma.dev compass-gateway-auth-oauth.local.kyma.dev compass-gateway.local.kyma.dev compass-gateway-int.local.kyma.dev compass.local.kyma.dev compass-mf.local.kyma.dev kyma-env-broker.local.kyma.dev director.local.kyma.dev compass-external-services-mock-sap-mtls.local.kyma.dev\" >> /etc/hosts"