#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="$( cd ${CURRENT_DIR}/../scripts && pwd )"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

DATA_DIR="${SCRIPTS_DIR}/../data"
mkdir -p "${DATA_DIR}"
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

# As of Kyma 2.6.3 we need to specify which namespaces should enable istio injection
RELEASE_NS=compass-system
kubectl create ns $RELEASE_NS --dry-run=client -o yaml | kubectl apply -f -
kubectl label ns $RELEASE_NS istio-injection=enabled --overwrite
# As of Kubernetes 1.25 we need to replace PodSecurityPolicies; we chose the Pod Security Standards
kubectl label ns $RELEASE_NS pod-security.kubernetes.io/enforce=baseline --overwrite

echo "Installing DB..."
helm upgrade --install --atomic --timeout "${TIMEOUT}" -f ./mergedOverrides.yaml --create-namespace --namespace "${RELEASE_NS}" localdb "${DB_CHARTS}"

echo "Look for DB dump in progress ..."
DBDUMP_JOB_JSON=$(kubectl get job compass-dbdump -n "${RELEASE_NS}" --ignore-not-found -o json)

if [ ! -z "$DBDUMP_JOB_JSON" ]; then
    echo "DB Dump job found. Wait to complete (up to 30 minutes) ..."
    kubectl wait --for=condition=complete job/compass-dbdump -n "${RELEASE_NS}" --timeout="${TIMEOUT}"
    DBDUMP_JOB_COMPLETION_STATUS=$(kubectl get job compass-dbdump -n "${RELEASE_NS}" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
    
    if [ "$DBDUMP_JOB_COMPLETION_STATUS" == "True" ]; then
        echo "DB Dump job was executed successfully. Cleanup cluster."
        kubectl delete job compass-dbdump -n "${RELEASE_NS}"
        kubectl delete serviceaccount dbdump-dbdump-job -n "${RELEASE_NS}"
    else
        echo "DB Dump job was failed. Exitting."
        DBDUMP_JOB_JSON=$(kubectl get job compass-dbdump -n "${RELEASE_NS}" --ignore-not-found -o json)
        echo "DB Dump job: $DBDUMP_JOB_JSON"
        exit 1
    fi
fi

trap "cleanup_trap" RETURN EXIT INT TERM
