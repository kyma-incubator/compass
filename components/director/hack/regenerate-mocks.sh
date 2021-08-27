#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )

echo "Generating mock implementation for interfaces..."
docker run --rm -v $PROJECT_ROOT:/home/app -w /home/app --entrypoint go vektra/mockery:latest -- generate ./...