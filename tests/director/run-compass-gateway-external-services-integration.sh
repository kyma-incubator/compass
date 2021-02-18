#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
COMPASS_PATH=$( cd "$( dirname "${ROOT_PATH}/../../../" )" && pwd )

suiteName="compass-gateway-external-services-integration"
nameSpace="kyma-system"

${COMPASS_PATH}/installation/scripts/testing-suite.sh ${nameSpace} ${suiteName}
