#!/usr/bin/env sh

# wait for Director to be up and running

echo -e "Checking if Director is up..."

if [ -z "$DIRECTOR_URL" ]
then
      echo "\$DIRECTOR_URL is empty"
      exit 1
fi

directorIsUp=false
set +e
for i in {1..10}; do
    curl --fail "${DIRECTOR_URL}/healthz"
    res=$?

    if [[ ${res} == 0 ]]; then
        directorIsUp=true
        break
    fi
    sleep 1
done
set -e

if [[ "$directorIsUp" == false ]]; then
    echo -e "Cannot access Director API"
    exit 1
fi