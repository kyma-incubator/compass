#!/usr/bin/env sh

# wait for Istio sidecar to be up and running

echo "Checking if Istio sidecar is running..."

if [ -z "$ISTIO_HEALTHZ_URL" ]; then
      echo "\$ISTIO_HEALTHZ_URL env variable is empty"
      exit 1
fi

i=0
maxRetries=${MAX_RETRIES:-30}
istioIsUp=false

set +e
while [ $i -lt "$maxRetries" ]
do
    curl --fail "${ISTIO_HEALTHZ_URL}"
    res=$?

    if [ "$res" -eq "0" ]; then
        istioIsUp=true
        break
    fi
    sleep 1
    i=$((i+1))
done

set -e

if [ "$istioIsUp" = false ]; then
    echo "Cannot access Istio sidecar proxy."
    exit 1
fi
