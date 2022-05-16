#!/usr/bin/env bash

make e2e-test
minikube ssh "sudo chmod -R 777 /examples"
USED_DRIVER=$(minikube profile list -o json | jq -r ".valid[0].Config.Driver")
if [[ $USED_DRIVER == "docker" ]]; then
  docker cp minikube:/examples/ ../../components/director/
else
  scp -r -i $(minikube ssh-key) docker@$(minikube ip):/examples/ ../../components/director/
fi
../../components/director/hack/prettify-examples.sh

