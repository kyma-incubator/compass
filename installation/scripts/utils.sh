#!/usr/bin/env bash

#
# Log is function useful for logs creation.
# It accepts three arguments:
# $1 - text we want to print
# $2 - text color
# $3 - text style
# It will create single log with defined color and font style.
# To specify style without color we have to put 'nc' before style.
# For example:
# log "gophers" magenta bold - will print bold 'gophers' in magenta color.
# log "text" bold - it will print normal text.
# log "text" nc bold - it will print bold text.
# By default log will print normal text like echo command.
# Use source [utils.sh path] to import log function into your script.
#

function log() {
    local exp=$1;
    local color=$2;
    local style=$3;
    local NC='\033[0m'
    if ! [[ ${color} =~ '^[0-9]$' ]] ; then
       case $(echo ${color} | tr '[:upper:]' '[:lower:]') in
        black) color='\e[30m' ;;
        red) color='\e[31m' ;;
        green) color='\e[32m' ;;
        yellow) color='\e[33m' ;;
        blue) color='\e[34m' ;;
        magenta) color='\e[35m' ;;
        cyan) color='\e[36m' ;;
        white) color='\e[37m' ;;
        nc|*) color=${NC} ;; # no color or invalid color
       esac
    fi
    if ! [[ ${style} =~ '^[0-9]$' ]] ; then
        case $(echo ${style} | tr '[:upper:]' '[:lower:]') in
        bold) style='\e[1m' ;;
        underline) style='\e[4m' ;;
        inverted) style='\e[7m' ;;
        *) style="" ;; # no style or invalid style
       esac
    fi
    printf "${color}${style}${exp}${NC}\n"
}

# checkInputParameterValue is a function to check if input parameter is valid
# There HAS to be provided argument:
# $1 - value for input parameter
# for example in installation/cmd/run.sh we can set --vm-driver argument, which has to have a value.

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}

function usek3d() {
    CURRENT_CONTEXT=$(kubectl config current-context)
    if [ $CURRENT_CONTEXT != "k3d-kyma"  ]; then
        echo "Current context is not 'k3d-kyma', switching to it..."
        kubectl config use-context k3d-kyma
        if [ $? -ne 0 ]; then
            echo "Failed to update context to 'k3d-kyma'. Local installation requires k3d running. To list all available contexts execute: 'kubectl config get-contexts'"
            return 1
        fi
    fi

    echo "Using 'k3d-kyma' kubectl context"
}

function is_helm_state_stable() {
    local DEPLOYMENT="$1"
    local NAMESPACE="$2"

    # Handle initial state when deployment will not be present 
    helm status ${DEPLOYMENT} -n ${NAMESPACE} > /dev/null 2>&1 || return 0

    STABLE_STATUSES=("deployed" "uninstalled" "superseded" "failed")
    CURRENT_STATUS="$(helm status ${DEPLOYMENT} -n ${NAMESPACE} -o json | jq -r .info.status)"
    echo "Current deployment status is: ${CURRENT_STATUS}" 
    for STATUS in "${STABLE_STATUSES[@]}"; do
        if [[ "${CURRENT_STATUS}" == "${STATUS}" ]]; then
            return 0
        fi
    done
    return 1
}

function wait_for_helm_stable_state() {
  local DEPLOYMENT="$1"
  local NAMESPACE="$2"
  START_TIME=$(date +%s)
  until is_helm_state_stable "${DEPLOYMENT}" "${NAMESPACE}" ; do
      CURRENT_TIME=$(date +%s)
      SECONDS=$((CURRENT_TIME-START_TIME))
      if (( SECONDS > 300 )); then
          echo "Timeout of 5 min for finishing installation of ${DEPLOYMENT} reached. Exiting."
          exit 1
      fi
      echo "Wait 10s before check again ..."
      sleep 10
  done
}

# Use kubectl with kubecontext for k3d cluster with name kyma
function kubectl_k3d_kyma () {
  kubectl --context k3d-kyma "$@"
}

# Use helm with kubecontext for k3d cluster with name kyma
function helm_k3d_kyma () {
  helm  --kube-context k3d-kyma "$@"
}

# Use kyma CLI with kubeconfig for k3d cluster with name kyma
# The Kyma CLI does not support --context
function kyma_k3d_kyma () {
  k3d kubeconfig get kyma > /tmp/k3d-kyma.yaml
  kyma --kubeconfig /tmp/k3d-kyma.yaml "$@"
}