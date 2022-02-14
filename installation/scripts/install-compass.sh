#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPASS_CHARTS="${CURRENT_DIR}/../../chart/compass"
COMPASS_OVERRIDES="${CURRENT_DIR}/../resources/helm-compass-overrides.yaml"
CRDS="${CURRENT_DIR}/../resources/crds"

kubectl apply -f "$CRDS"
helm install --wait --timeout 30m0s -f "$COMPASS_CHARTS"/values.yaml --create-namespace --namespace compass-system compass -f "$COMPASS_OVERRIDES" "$COMPASS_CHARTS"
