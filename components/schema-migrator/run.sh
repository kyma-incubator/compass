#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

CLEAN_DB='clean'

for var in DB_USER DB_HOST DB_NAME DB_PORT DB_PASSWORD MIGRATION_PATH DIRECTION; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

if [[ "${DIRECTION}" == "up" ]]; then
    echo "Migration UP"
elif [[ "${DIRECTION}" == "down" ]]; then
    echo "Migration DOWN"
else
    echo "ERROR: DIRECTION variable accepts only two values: up or down"
    exit 1
fi

echo '# WAITING FOR CONNECTION WITH DATABASE #'
for i in {1..30}
do
    pg_isready -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME"
    if [ $? -eq 0 ]
    then
        dbReady=true
        break
    fi
    sleep 1
done

if [ "${dbReady}" != true ] ; then
    echo '# COULD NOT ESTABLISH CONNECTION TO DATABASE #'
    exit 1
fi

DB_NAME_SSL=$DB_NAME
SSL_OPTION=""
if [ -n "${DB_SSL}" ] ; then
  DB_NAME_SSL="${DB_NAME}?sslmode=${DB_SSL}"
  SSL_OPTION="sslmode=${DB_SSL}"
fi


set -e

if [[ -f seeds/dump.sql ]] && [[ "${DIRECTION}" == "up" ]]; then
    echo "Will reuse existing dump in seeds/dump.sql"
    cat seeds/dump.sql | \
        PGPASSWORD="${DB_PASSWORD}" psql -v ON_ERROR_STOP=1 -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" --set="${SSL_OPTION}"

    REMOTE_MIGRATION_VERSION=$(PGPASSWORD="${DB_PASSWORD}" psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT version FROM schema_migrations")
    LOCAL_MIGRATION_VERSION=$(echo $(ls migrations/director | tail -n 1) | grep -o -E '[0-9]+' | head -1 | sed -e 's/^0\+//')

    if [[ ${REMOTE_MIGRATION_VERSION} = ${LOCAL_MIGRATION_VERSION} ]]; then
        echo -e "${GREEN}Both remote and local migrations are at the same version.${NC}"
    else
        echo -e "${RED}Remote and local migrations are at different versions.${NC}"
        echo -e "${YELLOW}REMOTE: $REMOTE_MIGRATION_VERSION${NC}"
        echo -e "${YELLOW}LOCAL: $LOCAL_MIGRATION_VERSION${NC}"
    fi
fi

CONNECTION_STRING="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME_SSL"
function currentVersion {
  echo $(migrate -path ${MIGRATION_STORAGE_PATH} -database "$CONNECTION_STRING" version 2>&1 | head -n1 | cut -d " " -f1)
}

LAST_SUCCESSFUL_MIGRATION=$(currentVersion)
echo "Last successful migration is $LAST_SUCCESSFUL_MIGRATION"
if [[ $LAST_SUCCESSFUL_MIGRATION == *'error'*]]; then
  LAST_SUCCESSFUL_MIGRATION=$CLEAN_DB
fi

CMD="migrate -path migrations/${MIGRATION_PATH} -database "$CONNECTION_STRING" ${DIRECTION}"

# validate.sh uses DRY_RUN
if [[ "${DRY_RUN}" == "true" ]]; then
      echo '# STARTING DRY-RUN MIGRATION #'
      yes | $CMD
      exit $?
fi

if [[ "${DIRECTION}" == "up" ]]; then
  # Save previous successful migration in case of reverting the Compass version
   kubectl create configmap -n ${CM_NAMESPACE} ${CM_NAME} --from-literal=version=$LAST_SUCCESSFUL_MIGRATION || \
      kubectl create configmap -n ${CM_NAMESPACE} ${CM_NAME} --from-literal=version=$LAST_SUCCESSFUL_MIGRATION --dry-run=client -o yaml | kubectl apply -f -

  echo "Replacing migrations in Persistent Volume with current migrations"
  rm  ${MIGRATION_STORAGE_PATH}/* || true
  cp -R migrations/${MIGRATION_PATH}/. ${MIGRATION_STORAGE_PATH}

  echo '# STARTING MIGRATION #'
  $CMD
else
  REVERT_TO=$(kubectl get configmap -n $CM_NAMESPACE $CM_NAME -o jsonpath='{.data.version}')
  echo "Will perform down migration to version $REVERT_TO"
  migrationExists=$(ls ${MIGRATION_STORAGE_PATH} | grep $REVERT_TO)
  if [[ -z $migrationExists ]] && [[ "$REVERT_TO" != "$CLEAN_DB" ]]; then
    echo "Migration version $REVERT_TO does not exist, please update the $CM_NAMESPACE/$CM_NAME configmap manually with the correct migration version. Available migrations are:"
    ls ${MIGRATION_STORAGE_PATH}
    exit 1
  else

  if [[ $LAST_SUCCESSFUL_MIGRATION != "$CLEAN_DB" ]]; then
    echo "Cleaning dirty flag"
    migrate -path ${MIGRATION_STORAGE_PATH} -database "$CONNECTION_STRING" force $LAST_SUCCESSFUL_MIGRATION
  fi

  # Migrate down until the version matches the wanted version from the previous release
  while [[ "$(currentVersion)" != "$REVERT_TO" ]]; do
    echo "Migrating down from $(currentVersion)"
    migrate -path ${MIGRATION_STORAGE_PATH} -database "$CONNECTION_STRING" down 1
  done

  echo "Successfully migrated down to previous migration version $REVERT_TO"
fi

set +e
