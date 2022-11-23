#!/bin/bash

###
# Following script installs necessary tooling for Debian, deploys Kyma with Compass on k3d, and runs the integrations tests.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

export ARTIFACTS="/var/log/prow_artifacts"
sudo mkdir -p "${ARTIFACTS}"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --dump-db)
            DUMP_DB=true
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


if [[ ${DUMP_DB} ]]; then
    sudo ${INSTALLATION_DIR}/cmd/run.sh --k3d-memory 12288MB --dump-db
else
    sudo ${INSTALLATION_DIR}/cmd/run.sh --k3d-memory 12288MB
fi

# Patch deployment imagePullPolicy to always pull image on subsequent executions of update-compass.sh
DEPLOYMENTS=$(kubectl get deployment -n compass-system | cut -d ' ' -f 1 | sed 1d)
namespace="compass-system"

for DEPLOYMENT in $DEPLOYMENTS; do
    if [ "$DEPLOYMENT" = "compass-pairing-adapter-adapter-local-mtls" ]; then
       	kubectl patch -n $namespace "deployment/$DEPLOYMENT" -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"pairing-adapter\",\"imagePullPolicy\":\"Always\"}]}}}}"
	    kubectl rollout restart -n $namespace "deployment/$DEPLOYMENT"
        continue
    fi
    CONTAINER_NAME=$(echo "$DEPLOYMENT" | cut -d '-' -f2-)
   	kubectl patch -n $namespace "deployment/$DEPLOYMENT" -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"$CONTAINER_NAME\",\"imagePullPolicy\":\"Always\"}]}}}}"
	kubectl rollout restart -n $namespace "deployment/$DEPLOYMENT"
done

kubectl patch -n $namespace cronjob/compass-ord-aggregator -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"ord-aggregator","imagePullPolicy":"Always"}]}}}}}}'
kubectl patch -n $namespace cronjob/compass-system-fetcher -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"system-fetcher","imagePullPolicy":"Always"}]}}}}}}'
kubectl patch -n $namespace cronjob/compass-director-tenant-loader-external -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"loader","imagePullPolicy":"Always"}]}}}}}}'

