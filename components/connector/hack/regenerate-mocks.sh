#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(dirname ${BASH_SOURCE})/..

echo "Installing mockery 2.9.0..."
go install github.com/vektra/mockery/v2/.../@v2.9.0
echo "Installing latest failery..."
go install github.com/kyma-project/kyma/tools/failery/.../
echo "Generating mock implementation for interfaces..."
cd ${PROJECT_ROOT}
go generate ./...
