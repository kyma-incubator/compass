#!/usr/bin/env sh

# wait for Director to be up and running

echo "Checking if Director is up..."

if [ -z "$DIRECTOR_HEALTHZ_URL" ]; then
      echo "\$DIRECTOR_HEALTHZ_URL env variable is empty"
      exit 1
fi

i=0
maxRetries=${MAX_RETRIES:-30}
directorIsUp=false

set +e
while [ $i -lt "$maxRetries" ]
do
    curl --fail "${DIRECTOR_HEALTHZ_URL}"
    res=$?

    if [ "$res" -eq "0" ]; then
        directorIsUp=true
        break
    fi
    sleep 1
    true $(( i++ ))
done

set -e

if [ "$directorIsUp" = false ]; then
    echo "Cannot access Director API"
    exit 1
fi
