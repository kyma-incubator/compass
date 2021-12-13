#!/usr/bin/env bash

make e2e-test
minikube ssh "sudo chmod -R 777 /examples"
scp -r -i $(minikube ssh-key) docker@$(minikube ip):/examples/ ../../components/director/
../../components/director/hack/prettify-examples.sh

