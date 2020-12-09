#!/usr/bin/env sh

# wait for Tenant Fetcher to be up and running

echo "Checking if Tenant Fetcher is up..."

if [ -z "$TENANT_FETCHER_HEALTHZ_URL" ]; then
      echo "\$TENANT_FETCHER_HEALTHZ_URL env variable is empty"
      exit 1
fi

i=0
maxRetries=${MAX_RETRIES:-60}
tenantFetcherIsUp=false

set +e
while [ $i -lt "$maxRetries" ]
do
    curl --fail "${TENANT_FETCHER_HEALTHZ_URL}"
    res=$?

    if [ "$res" -eq "0" ]; then
        tenantFetcherIsUp=true
        break
    fi
    sleep 1
    i=$((i+1))
done

set -e

if [ "$tenantFetcherIsUp" = false ]; then
    echo "Cannot access Tenant Fetcher API"
    exit 1
fi
