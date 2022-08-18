#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

DB_CHARTS="${CURRENT_DIR}/../../chart/postgresql"

COMPASS_CHARTS="${CURRENT_DIR}/../../chart/compass"

function cleanup_trap() {
  if [[ -f mergedOverrides.yaml ]]; then
    rm -f mergedOverrides.yaml
  fi
}

touch mergedOverrides.yaml # target file where all overrides .yaml files will be merged into. This is needed because if several override files with the same key/s are passed to helm, it applies the value/s from the last file for that key overriding everything else.
yq eval-all --inplace '. as $item ireduce ({}; . * $item )' mergedOverrides.yaml "${COMPASS_CHARTS}"/values.yaml

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

echo "Install DB"
helm upgrade --install --wait --debug --timeout "${TIMEOUT}" -f ./mergedOverrides.yaml --create-namespace --namespace compass-system postgresql "${DB_CHARTS}"

echo "Look for DB migrations in progress ..."
MIGRATION_JOB_JSON=$(kubectl get job compass-migration -n compass-system --ignore-not-found -o json)

if [ ! -z "$MIGRATION_JOB_JSON" ]; then
    echo "Migration job found. Wait to complete (up to 30 minutes) ..."
    kubectl wait --for=condition=complete job/compass-migration -n compass-system --timeout="${TIMEOUT}"
    MIGRATION_JOB_COMPLETION_STATUS=$(kubectl get job compass-migration -n compass-system -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
    
    if [ "$MIGRATION_JOB_COMPLETION_STATUS" == "True" ]; then
        echo "Migration job was executed successfully. Cleanup cluster."
        kubectl delete job compass-migration -n compass-system
        kubectl delete serviceaccount postgresql-migrator-job -n compass-system
        kubectl delete persistentvolumeclaim compass-director-migrations -n compass-system
    else
        echo "Migration job was failed. Exitting."
        MIGRATION_JOB_JSON=$(kubectl get job compass-migration -n compass-system --ignore-not-found -o json)
        echo "Migration job: $MIGRATION_JOB_JSON"
        exit 1
    fi
fi

echo "Install Compass"
CRDS_FOLDER="${CURRENT_DIR}/../resources/crds"
kubectl apply -f "${CRDS_FOLDER}"
helm upgrade --install --wait --debug --timeout "${TIMEOUT}" -f ./mergedOverrides.yaml --create-namespace --namespace compass-system compass "${COMPASS_CHARTS}"
trap "cleanup_trap" RETURN EXIT INT TERM