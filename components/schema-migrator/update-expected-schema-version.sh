#!/usr/bin/env bash
MIGRATION_PATH="${1}"
CONFIGMAP_NAME="${2}"
LATEST_SCHEMA=$(ls -lr migrations/${MIGRATION_PATH} | head -n 2 | tail -n 1 | tr -s ' ' | cut -d ' ' -f9 | cut -d '_' -f1)

kubectl get -n compass-system configmap ${CONFIGMAP_NAME} -o json | \
jq --arg e "$LATEST_SCHEMA"  '.data.schemaVersion = $e'  > temp_configmap.json
kubectl apply -f temp_configmap.json
rm temp_configmap.json