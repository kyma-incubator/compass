#!/usr/bin/env bash

COMPONENT=$1
touch td_duplicate.yaml
kubectl get testdefinition compass-e2e-$COMPONENT -n kyma-system -o json > td_duplicate.json
export TD_NAME=compass-e2e-$COMPONENT-local
jq --arg e "$TD_NAME"  '.metadata.name = $e' td_duplicate.json |sponge td_duplicate.json
jq '.spec.template.spec.containers[0].args[1] = "sleep 99999"' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.creationTimestamp)' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.generation)' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.selfLink)' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.uid)' td_duplicate.json |sponge td_duplicate.json

kubectl apply -f td_duplicate.json

sed "s/PLACEHOLDER/$COMPONENT-local/" test-suite.yaml | kubectl -n kyma-system apply -f -

kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0 -c $COMPONENT-tests -- apk add go code=$?
while code != 0
do
  sleep 2; kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0 -c $COMPONENT-tests -- apk add go code=$?
done

kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0 -c $COMPONENT-tests -- apk add go
kubectl cp . kyma-system/oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0:tests -c $COMPONENT-tests