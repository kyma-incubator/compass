#!/usr/bin/env bash

# This script is responsible for running tests for Director.

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

SCOPES_CONFIGURATION_FILE=${ROOT_PATH}/chart/compass/charts/director/scopes.yaml go test ./...