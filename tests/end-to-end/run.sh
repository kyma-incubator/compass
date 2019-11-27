#!/usr/bin/env bash

# This script is responsible for running tests for Director.

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

ALL_SCOPES="runtime:write application:write label_definition:write integration_system:write application:read runtime:read label_definition:read integration_system:read health_checks:read application_template:read application_template:write" go test ./...
