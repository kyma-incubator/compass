#!/usr/bin/env bash

COMPONENT=$1
echo $COMPONENT
kubectl get testdefinition compass-e2e-$COMPONENT -n kyma-system -o json > td_duplicate.json
export TD_NAME=compass-e2e-$COMPONENT-local
jq --arg e "$TD_NAME"  '.metadata.name = $e' td_duplicate.json |sponge td_duplicate.json
jq '.spec.template.spec.containers[0].args[1] = "sleep 99999"' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.creationTimestamp)' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.generation)' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.selfLink)' td_duplicate.json |sponge td_duplicate.json
jq 'del(.metadata.uid)' td_duplicate.json |sponge td_duplicate.json

kubectl apply -f td_duplicate.json
rm td_duplicate.json

cat <<EOF | sed "s/PLACEHOLDER/$COMPONENT-local/" | kubectl -n kyma-system apply -f -
apiVersion: testing.kyma-project.io/v1alpha1
kind: ClusterTestSuite
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: compass-e2e-tests
spec:
  maxRetries: 0
  concurrency: 1
  selectors:
    matchNames:
      - name: compass-e2e-PLACEHOLDER
        namespace: kyma-system
EOF

containerStatus=$(kubectl get pods -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0  -ojsonpath="{.status.containerStatuses[?(@.name=='$COMPONENT-tests')].state.running}")

while [ -z "$containerStatus" ]
do
    sleep 2
    containerStatus=$(kubectl get pods -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0  -ojsonpath="{.status.containerStatuses[?(@.name=='$COMPONENT-tests')].state.running}")
    echo "Container not started. Waiting..."
done

kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-$COMPONENT-local-0 -c $COMPONENT-tests -- apk add go