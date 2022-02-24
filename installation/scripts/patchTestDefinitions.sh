#!/usr/bin/env bash

IMAGE_VERSION=$1
for td in $(kubectl get testdefinition -n kyma-system -o name); do
   echo "${td}"
  container=$(kubectl get "${td}" -n kyma-system -o=jsonpath='{$.spec.template.spec.containers[0].name}')
  echo "${container}"
  kubectl -n kyma-system patch --type merge "${td}" \
  		-p '{"spec":{"template":{"spec":{"containers":[{"name":"'"${container}"'","image":"k3d-kyma-registry:5001/compass-tests:'"${IMAGE_VERSION}"'"}]}}}}'
done