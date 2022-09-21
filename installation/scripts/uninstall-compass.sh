#!/bin/bash

###
# Following script uninstalls compass installation only.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="${CURRENT_DIR}/../scripts"
source $SCRIPTS_DIR/utils.sh

TIMEOUT=30m0s

echo "Wait for helm stable status"
wait_for_helm_stable_state "compass" "compass-system" 

echo "Uninstall Compass"
helm uninstall --wait --debug --timeout "${TIMEOUT}" --namespace compass-system compass || true
