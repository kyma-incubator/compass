#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="$( cd ${CURRENT_DIR}/../scripts && pwd )"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

DATA_DIR="$( cd ${SCRIPTS_DIR}/../data && pwd )"
DB_CHARTS="$( cd ${CURRENT_DIR}/../../chart/localdb && pwd )"

function cleanup_trap() {
  if [[ -f mergedOverrides.yaml ]]; then
    rm -f mergedOverrides.yaml
  fi
}

touch mergedOverrides.yaml # target file where all overrides .yaml files will be merged into. This is needed because if several override files with the same key/s are passed to helm, it applies the value/s from the last file for that key overriding everything else.
yq eval-all --inplace '. as $item ireduce ({}; . * $item )' mergedOverrides.yaml "${DB_CHARTS}"/values.yaml

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --overrides-file)
            checkInputParameterValue "${2}"
            yq eval-all --inplace '. as $item ireduce ({}; . * $item )' mergedOverrides.yaml ${2}
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

echo "Wait for helm stable status"
wait_for_helm_stable_state "localdb" "compass-system" 

echo "Install DB"
helm upgrade --install --wait --debug --timeout "${TIMEOUT}" -f ./mergedOverrides.yaml --create-namespace --namespace compass-system localdb "${DB_CHARTS}"

echo "Look for DB dump in progress ..."
DBDUMP_JOB_JSON=$(kubectl get job compass-dbdump -n compass-system --ignore-not-found -o json)

if [ ! -z "$DBDUMP_JOB_JSON" ]; then
    echo "DB Dump job found. Wait to complete (up to 30 minutes) ..."
    kubectl wait --for=condition=complete job/compass-dbdump -n compass-system --timeout="${TIMEOUT}"
    DBDUMP_JOB_COMPLETION_STATUS=$(kubectl get job compass-dbdump -n compass-system -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
    
    if [ "$DBDUMP_JOB_COMPLETION_STATUS" == "True" ]; then
        echo "DB Dump job was executed successfully. Cleanup cluster."
        kubectl delete job compass-dbdump -n compass-system
        kubectl delete serviceaccount dbdump-dbdump-job -n compass-system
    else
        echo "DB Dump job was failed. Exitting."
        DBDUMP_JOB_JSON=$(kubectl get job compass-dbdump -n compass-system --ignore-not-found -o json)
        echo "DB Dump job: $DBDUMP_JOB_JSON"
        exit 1
    fi
fi

trap "cleanup_trap" RETURN EXIT INT TERM