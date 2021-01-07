#!/usr/bin/env sh

# wait for Director and ORD Service to be up and running

ping() {
   COMPONENT_NAME=$1
   COMPONENT_HEALTHZ_URL=$2

   echo "Checking if $COMPONENT_NAME is up..."

   if [ -z "$COMPONENT_HEALTHZ_URL" ]; then
         echo "\$COMPONENT_HEALTHZ_URL env variable is empty"
         exit 1
   fi

   i=0
   maxRetries=${MAX_RETRIES:-60}
   componentIsUp=false

   set +e
   while [ $i -lt "$maxRetries" ]
   do
       curl --fail "${COMPONENT_HEALTHZ_URL}"
       res=$?

       if [ "$res" -eq "0" ]; then
           componentIsUp=true
           break
       fi
       sleep 1
       i=$((i+1))
   done

   set -e

   if [ "$componentIsUp" = false ]; then
       echo "Cannot access $COMPONENT_NAME API"
       exit 1
   fi
}

ping "Director" $DIRECTOR_HEALTHZ_URL
ping "System Broker" $SYSTEM_BROKER_HEALTHZ_URL
