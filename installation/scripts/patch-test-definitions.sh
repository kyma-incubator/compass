#!/usr/bin/env bash

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$CURRENT_DIR"/utils.sh

IMAGE_VERSION=$1
for td in $(kubectl_k3d_kyma get testdefinition -n kyma-system -o name); do
  TEST_DEFINITION=$(kubectl_k3d_kyma get -n kyma-system "${td}" -o yaml)
  echo "${TEST_DEFINITION}" | sed 's/image\:.*/image\: k3d-kyma-registry\:5001\/compass-e2e-tests\:'"${IMAGE_VERSION}"'/' | sed 's/imagePullPolicy\:.*/imagePullPolicy\: Always/' | kubectl_k3d_kyma apply -f - 2>/dev/null
done
