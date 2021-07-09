#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(dirname ${BASH_SOURCE})/..

echo "Installing mockery 2.9.0..."
go get github.com/vektra/mockery/v2/.../@v2.9.0
echo "Generating mock implementation for interfaces..."
cd ${PROJECT_ROOT}
go generate ./...
