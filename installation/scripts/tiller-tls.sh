#!/usr/bin/env bash

RETRY_TIME=5
MAX_RETRIES=3
SECRET_NAME="helm-secret"
NAMESPACE="compass-installer"
NAMESPACE_FALLBACK="kyma-integration"

mkdir -p "$(helm home)"

function findHelmSecret() {
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" > /dev/null
}

function findHelmSecretFallback() {
    kubectl get -n "${NAMESPACE_FALLBACK}" secret "${SECRET_NAME}" > /dev/null
}

function copyHelmSecret() {
    kubectl get secret "${SECRET_NAME}" --namespace="${NAMESPACE_FALLBACK}" --export -o yaml | kubectl apply --namespace="${NAMESPACE}" -f -
}

function defer() {
    local current="${1}"
    if [[ "${current}" -eq "${MAX_RETRIES}" ]]; then return 1; fi
    echo "---> Retrying in ${RETRY_TIME} seconds..."
    sleep "${RETRY_TIME}"
}

function fail() {
    echo "---> Warning! Unable to find Helm secret: timeout."
    exit 1
}

function saveCerts {
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.ca\\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.tls\\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
    kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.tls\\.key']}" | base64 --decode > "$(helm home)/key.pem"
}

echo "---> Finding Helm secret..."
for i in $(seq 1 "${MAX_RETRIES}"); do (findHelmSecret || (findHelmSecretFallback && copyHelmSecret)) && break || defer "${i}" || fail ; done

echo "---> Helm secret found. Saving Helm certificates under the \"$(helm home)\" directory..."
saveCerts