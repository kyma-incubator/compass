#!/bin/bash

export OIDC_KUBECONFIG_ISSUER_URL=http://some-issuer.for.kubernetes.io 
export OIDC_KUBECONFIG_CLIENT_ID=some-id-for-k8s
export OIDC_KUBECONFIG_CLIENT_SECRET=super-strong-password
export OIDC_ISSUER_URL=https://some-issuer.for.token-validation
export OIDC_CLIENT_ID=some-id-for-token
export OIDC_CLIENT_SECRET=also-strong-password
go run cmd/generator/main.go