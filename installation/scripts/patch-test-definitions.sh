#!/usr/bin/env bash

IMAGE_VERSION=$1
for td in $(kubectl get testdefinition -n kyma-system -o name); do
  TEST_DEFINITION=$(kubectl get -n kyma-system "${td}" -o yaml)
  echo "${TEST_DEFINITION}" | sed 's/image\:.*/image\: k3d-kyma-registry\:5001\/compass-e2e-tests\:'"${IMAGE_VERSION}"'/' | sed 's/imagePullPolicy\:.*/imagePullPolicy\: Always/' | kubectl apply -f - 2>/dev/null
done
