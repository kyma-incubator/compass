#!/usr/bin/env bash

set -o errexit

GREEN='\033[0;32m'
NC='\033[0m' # No Color

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$CURRENT_DIR"/utils.sh

TIMEOUT=30m0s

COMPASS_CHARTS="${CURRENT_DIR}/../../chart/compass"

# $KUBECTL will take on the value "kubectl" if not overridden - in local installation `run.sh` will override it
# This is done to make sure that `install-compass.sh` can be used with different kubeconfigs and not only the local one
: ${KUBECTL:=kubectl}
# $HELM will take on the value "helm" if not overridden - in local installation `run.sh` will override it
# This is done to make sure that `install-compass.sh` can be used with different kubeconfigs and not only the local one
: ${HELM:=helm}

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
        --compass-charts)
            checkInputParameterValue "${2}"
            COMPASS_CHARTS="${2}"
            shift # past argument
            shift
        ;;
        --sql-helm-backend)
            SQL_HELM_BACKEND=true
            shift # past argument
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
"$KUBECTL" create ns $RELEASE_NS --dry-run=client -o yaml | "$KUBECTL" apply -f -
"$KUBECTL" label ns $RELEASE_NS istio-injection=enabled --overwrite
# As of Kubernetes 1.25 we need to replace PodSecurityPolicies; we chose the Pod Security Standards
"$KUBECTL" label ns $RELEASE_NS pod-security.kubernetes.io/enforce=baseline --overwrite

if [[ ${SQL_HELM_BACKEND} ]]; then
    echo -e "${GREEN}Helm SQL storage backend will be used${NC}"

    DB_USER=$(base64 -d <<< $("$KUBECTL" get secret -n "${RELEASE_NS}" compass-postgresql -o=jsonpath="{.data['postgresql-director-username']}"))
    DB_PWD=$(base64 -d <<< $("$KUBECTL" get secret -n "${RELEASE_NS}" compass-postgresql -o=jsonpath="{.data['postgresql-director-password']}"))
    DB_NAME=$(base64 -d <<< $("$KUBECTL" get secret -n "${RELEASE_NS}" compass-postgresql -o=jsonpath="{.data['postgresql-directorDatabaseName']}"))
    DB_PORT=$(base64 -d <<< $("$KUBECTL" get secret -n "${RELEASE_NS}" compass-postgresql -o=jsonpath="{.data['postgresql-servicePort']}"))

    "$KUBECTL" port-forward --namespace "${RELEASE_NS}" svc/compass-postgresql ${DB_PORT}:${DB_PORT} &
    sleep 5 #wait port-forwarding to be completed

    export HELM_DRIVER=sql
    export HELM_DRIVER_SQL_CONNECTION_STRING=postgres://${DB_USER}:${DB_PWD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable
fi

echo "Wait for helm stable status..."
wait_for_helm_stable_state "compass" ""${RELEASE_NS}"" 

echo "Starting compass installation..."
echo "Path to compass charts: " ${COMPASS_CHARTS}
"$HELM" upgrade --install --atomic --timeout "${TIMEOUT}" -f ./mergedOverrides.yaml --create-namespace --namespace "${RELEASE_NS}" compass "${COMPASS_CHARTS}"
trap "cleanup_trap" RETURN EXIT INT TERM
echo "Compass installation finished successfully"

STATUS=$("$HELM" status compass -n compass-system -o json | jq .info.status)
echo "Compass installation status ${STATUS}"

if [[ $(uname -m) == "arm64" ]]; then
  echo "Patching image on octopus for arm64..."
 	"$KUBECTL" set image -n "${RELEASE_NS}" statefulset/compass-octopus "manager=europe-west1-docker.pkg.dev/sap-cp-cmp-dev/ucl-dev/octopus:5f353cd5"
fi

if [[ ${SQL_HELM_BACKEND} ]]; then
    pkill kubectl
fi
