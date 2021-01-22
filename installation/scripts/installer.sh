#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"
INSTALLER="${RESOURCES_DIR}/installer-local.yaml"
INSTALLER_CONFIG="${RESOURCES_DIR}/installer-config-local.yaml.tpl"
INSTALLER_CONFIG_KYMA_OVERRIDES="${RESOURCES_DIR}/installer-config-local-compass-with-full-kyma.yaml"
AZURE_BROKER_CONFIG=""
HELM_VERSION=$(helm version --short -c | cut -d '.' -f 1)

source $CURRENT_DIR/utils.sh

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    
    key="$1"

    case ${key} in
        --cr)
            checkInputParameterValue "$2"
            CR_PATH="$2"
            shift # past argument
            shift # past value
            ;;
        --password)
            ADMIN_PASSWORD="$2"
            shift
            shift
            ;;
         --kyma-installation)
            checkInputParameterValue "${2}"
            KYMA_INSTALLATION="$2"
            shift
            shift
            ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *) # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo "
################################################################################
# Compass Installer setup
################################################################################
"

bash ${CURRENT_DIR}/is-ready.sh kube-system k8s-app kube-dns

if [ $CR_PATH ]; then

    case $CR_PATH in
    /*) ;;
    *) CR_PATH="$(pwd)/$CR_PATH";;
    esac

    if [ ! -f $CR_PATH ]; then
        echo "CR file not found in path $CR_PATH"
        exit 1
    fi

fi

echo -e "\nCreating installation combo yaml"
if [[ $KYMA_INSTALLATION == *full* ]]; then
  echo -e "Preparing combo installer for compass with full kyma"
  COMBO_YAML=$(bash ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER} ${INSTALLER_CONFIG} ${INSTALLER_CONFIG_KYMA_OVERRIDES} ${AZURE_BROKER_CONFIG})
else
  echo -e "Preparing combo installer for compass with minimal kyma"
  COMBO_YAML=$(bash ${CURRENT_DIR}/concat-yamls.sh ${INSTALLER} ${INSTALLER_CONFIG} ${AZURE_BROKER_CONFIG})
fi


rm -rf ${AZURE_BROKER_CONFIG}

if [ ${ADMIN_PASSWORD} ]; then
    ADMIN_PASSWORD=$(echo ${ADMIN_PASSWORD} | tr -d '\n' | base64)
    COMBO_YAML=$(sed 's/global\.adminPassword: .*/global.adminPassword: '"${ADMIN_PASSWORD}"'/g' <<<"$COMBO_YAML")
fi

MINIKUBE_IP=$(minikube ip)
COMBO_YAML=$(sed 's/\.minikubeIP: .*/\.minikubeIP: '"${MINIKUBE_IP}"'/g' <<<"$COMBO_YAML")

echo -e "\nConfiguring sub-components"
bash ${CURRENT_DIR}/configure-components.sh

echo -e "\nStarting installation!"
kubectl apply -f - <<< "$COMBO_YAML"
kubectl rollout restart deployment -n compass-installer compass-installer
sleep 15
kubectl apply -f "${CR_PATH}"
