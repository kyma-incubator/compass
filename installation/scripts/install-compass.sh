#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --overrides-file)
            checkInputParameterValue "${2}"
            COMPASS_OVERRIDES="${2}"
            shift # past argument
            shift
        ;;
        --timeout)
            checkInputParameterValue "${2}"
            TIMEOUT="${2}"
            shift # past argument
            shift
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

COMPASS_CHARTS="${CURRENT_DIR}/../../chart/compass"
CRDS_FOLDER="${CURRENT_DIR}/../resources/crds"
kubectl apply -f "${CRDS_FOLDER}"
helm install --wait --timeout "${TIMEOUT}" -f "${COMPASS_CHARTS}"/values.yaml --create-namespace --namespace compass-system compass -f "${COMPASS_OVERRIDES}" "${COMPASS_CHARTS}"
