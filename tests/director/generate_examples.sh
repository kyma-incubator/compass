#!/usr/bin/env bash

minikube mount ../../components/director/examples:/examples &
trap "kill -9 $!" EXIT
sleep 1
sed "s/PLACEHOLDER/director/" ../test-suite.yaml | kubectl -n kyma-system apply -f -
while true
do
    statusSucceeded=$(kubectl get cts compass-e2e-tests -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$(kubectl get cts compass-e2e-tests -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$(kubectl get cts compass-e2e-tests -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

    if [[ "${statusSucceeded}" == *"True"* ]]; then
       echo "Test suite 'compass-e2e-tests' succeeded."
       break
    fi

    if [[ "${statusFailed}" == *"True"* ]]; then
        echo "Test suite 'compass-e2e-tests' failed."
        break
    fi

    if [[ "${statusError}" == *"True"* ]]; then
        echo "Test suite 'compass-e2e-tests' errored."
        break
    fi

    echo "ClusterTestSuite not finished. Waiting..."
    sleep 3
done

