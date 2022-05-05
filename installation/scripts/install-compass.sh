#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

COMPASS_CHARTS="${CURRENT_DIR}/../../chart/compass"

touch mergedOverrides.yaml # target file where all overrides .yaml files will be merged into. This is needed because if several override files with the same key/s are passed to helm, it applies the value/s from the last file for that key overriding everything else.
yq merge --inplace --overwrite mergedOverrides.yaml "${COMPASS_CHARTS}"/values.yaml

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --overrides-file)
            checkInputParameterValue "${2}"
            yq merge --inplace --overwrite mergedOverrides.yaml ${2}
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

CRDS_FOLDER="${CURRENT_DIR}/../resources/crds"
kubectl apply -f "${CRDS_FOLDER}"
helm install --wait --timeout "${TIMEOUT}" -f ./mergedOverrides.yaml --create-namespace --namespace compass-system compass "${COMPASS_CHARTS}"
rm mergedOverrides.yaml