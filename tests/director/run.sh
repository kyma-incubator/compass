#!/usr/bin/env bash

# This script is responsible for running tests for Director.

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

set -o errexit

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../..

source "${ROOT_PATH}/tests/director/common.sh"

DOMAIN="kyma.local" \
GATEWAY_OAUTH20_SUBDOMAIN="compass-gateway-auth-oauth" \
GATEWAY_JWT_SUBDOMAIN="compass-gateway" \
GATEWAY_CLIENT_CERTS_SUBDOMAIN="compass-gateway-mtls" \
DEFAULT_TENANT="3e64ebae-38b5-46a0-b1ed-9ccee153a0ae" \
go test -v ./...
