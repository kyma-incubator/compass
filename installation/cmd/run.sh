#!/usr/bin/env bash

set -o errexit

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

MINIKUBE_MEMORY=8192
MINIKUBE_TIMEOUT=25m
MINIKUBE_CPUS=5

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

if [ -z "$KYMA_RELEASE" ]; then
  KYMA_RELEASE=$(<"${ROOT_PATH}"/installation/resources/KYMA_VERSION)
fi

if [ -z "$KYMA_INSTALLATION" ]; then
  KYMA_INSTALLATION="minimal"
fi

if [[ ! ${SKIP_MINIKUBE_START} ]]; then
  echo "Provisioning Minikube cluster..."
  if [[ ! ${DOCKER_DRIVER} ]]; then
    kyma provision minikube --cpus ${MINIKUBE_CPUS} --memory ${MINIKUBE_MEMORY} --timeout ${MINIKUBE_TIMEOUT}
  else
    kyma provision minikube --cpus ${MINIKUBE_CPUS} --memory ${MINIKUBE_MEMORY} --timeout ${MINIKUBE_TIMEOUT} --vm-driver docker --docker-ports 443:443 --docker-ports 80:80
  fi
fi

if [[ ! ${SKIP_KYMA_START} ]]; then
  LOCAL_ENV=true bash "${ROOT_PATH}"/installation/scripts/install-kyma.sh --kyma-release ${KYMA_RELEASE} --kyma-installation ${KYMA_INSTALLATION}
fi

bash "${ROOT_PATH}"/installation/scripts/run-compass-installer.sh --kyma-installation ${KYMA_INSTALLATION}
bash "${ROOT_PATH}"/installation/scripts/is-installed.sh

echo "Adding Compass entries to /etc/hosts..."
sudo sh -c 'echo "\n$(minikube ip) adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local" >> /etc/hosts'