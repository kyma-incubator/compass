#!/usr/bin/env bash

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}

while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --component)
            checkInputParameterValue "${2}"
            MIGRATION_PATH="${2}"
            shift # past argument
            shift # past value
        ;;
        --cm-name)
            checkInputParameterValue "${2}"
            CONFIGMAP_NAME="${2}"
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

LATEST_SCHEMA=$(ls migrations/${MIGRATION_PATH} | tail -n 1 | grep -o -E '^[0-9]+' | sed -e 's/^0\+//')

kubectl get -n compass-system configmap ${CONFIGMAP_NAME} -o json | \
jq --arg e "$LATEST_SCHEMA"  '.data.schemaVersion = $e'  > temp_configmap.json
kubectl apply -f temp_configmap.json
rm temp_configmap.json
