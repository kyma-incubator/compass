#!/bin/bash
export GO111MODULE=on

ROOT_PATH=$(dirname "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)")

KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-incubator"}
JOB_NAME_PATTERN=${JOB_NAME_PATTERN:-"(pull-.*)"}
TIMEOUT=${JOBGUARD_TIMEOUT:-"15m"}

if [ -z "$PULL_PULL_SHA" ]; then
  echo "WORKAROUND: skip jobguard execution - not on PR commit"
  exit 0
fi

args=(
  "-github-endpoint=http://ghproxy"
  "-github-endpoint=https://api.github.com"
  "-github-token-path=/etc/github/token"
  "-fail-on-no-contexts=false"
  "-timeout=$TIMEOUT"
  "-org=$REPO_OWNER"
  "-repo=$REPO_NAME"
  "-base-ref=$PULL_PULL_SHA"
  "-expected-contexts-regexp=$JOB_NAME_PATTERN"
)

if [ -x "/prow-tools/jobguard" ]; then
  /prow-tools/jobguard "${args[@]}"
else
  cd "${ROOT_PATH}/installation/scripts/kyma-scripts/jobguard/cmd/jobguard" || exit 1
  go run main.go "${args[@]}"
fi
