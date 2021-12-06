#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )

echo "Cleaning up old mocks"
find . -name automock -type d -exec rm -r "{}" \; || true

echo "Generating new mock implementation for interfaces..."
docker image rm vektra/mockery:latest # This is needed to ensure that the new latest image will be downloaded and the version will be the same for everyone
docker run --rm -v $PROJECT_ROOT:/home/app -w /home/app --entrypoint go vektra/mockery:latest -- generate ./...