#!/bin/bash

###
# Following script uninstalls compass installation only.
#

set -o errexit

TIMEOUT=30m0s

echo "Uninstall Compass"
helm uninstall --wait --debug --timeout "${TIMEOUT}" --namespace compass-system compass

# TODO check if DOWN migrations are also executed