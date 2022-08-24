#!/usr/bin/env bash

make e2e-test
docker cp k3d-kyma-agent-0:/examples/ ../../components/director/

if [[ $? != 0 ]]
then
  echo "Searching examples in k3d-kyma-server-0..."
  docker cp k3d-kyma-server-0:/examples/ ../../components/director/
fi

../../components/director/hack/prettify-examples.sh