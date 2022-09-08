#!/bin/bash

###
# Following script uninstalls compass installation.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

sudo ${INSTALLATION_DIR}/scripts/uninstall-compass.sh
