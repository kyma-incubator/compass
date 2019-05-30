#!/usr/bin/env bash

kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem";

helm install --namespace='kyma-system' --name='compass' --tls --values ./compass/values.yaml ./compass/

sudo sh -c 'echo "$(minikube ip) compass-gateway.kyma.local" >> /etc/hosts'
