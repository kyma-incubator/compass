#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

CLEAN_DB_MESSAGE='error: no migration'

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}
//
POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --component)
            checkInputParameterValue "${2}"
            MIGRATION_PATH="${2}"
            shift # past argument
            shift # past value
        ;;
        --pv-path)
            checkInputParameterValue "${2}"
            MIGRATION_STORAGE_PATH="${2}"
            shift # past argument
            shift # past value
        ;;
        --up)
            DIRECTION=up
            shift # past argument
        ;;
        --down)
            DIRECTION=down
            shift # past argument
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

for var in DB_USER DB_HOST DB_NAME DB_PORT DB_PASSWORD MIGRATION_PATH DIRECTION; do
    if [ -z "${!var}" ]; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [[ -z "$DRY_RUN" ]] && [[ -z "$MIGRATION_STORAGE_PATH" ]] ; then
  echo "ERROR: MIGRATION_STORAGE_PATH is not set"
  discoverUnsetVar=true
fi

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

CONNECTION_STRING="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME_SSL"
function currentVersion {
  echo $(migrate -path ${MIGRATION_STORAGE_PATH} -database "$CONNECTION_STRING" version 2>&1)
}

function ensureMigrationExists() {
    LOCATION=$1
    MIGRATION=$2
    echo "Ensuring migration version \"$MIGRATION\" exists in \"$LOCATION\"..."
    if [[ -z "$MIGRATION" ]] || [[ -z $(ls -al "$LOCATION" | grep "$MIGRATION") ]]; then
      echo "Migration version \"$MIGRATION\" does not exist in \"$LOCATION\". Available migrations are:"
      ls -al $LOCATION || true
      exit 1
    fi
}

CMD="migrate -path migrations/${MIGRATION_PATH} -database "$CONNECTION_STRING" ${DIRECTION}"

# validate.sh uses DRY_RUN
if [[ "${DRY_RUN}" == "true" ]]; then
      echo '# STARTING DRY-RUN MIGRATION #'
      yes | $CMD
      exit $?
fi

if [[ "${DIRECTION}" == "up" ]]; then
  echo "Replacing migrations in Persistent Volume with current migrations"
  rm  ${MIGRATION_STORAGE_PATH}/* || true
  cp -R migrations/${MIGRATION_PATH}/. ${MIGRATION_STORAGE_PATH}

  PREVIOUS_MIGRATIONS_EXISTS=$(PGPASSWORD="${DB_PASSWORD}" psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename  = 'schema_migrations');")
  if [[ "$PREVIOUS_MIGRATIONS_EXISTS" == t ]]; then
    echo "Previous migrations exist, verifying that the state is not dirty."
    DATABASE_STATE=$(PGPASSWORD="${DB_PASSWORD}" psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT dirty FROM schema_migrations")
    echo "DB state is dirty: $DATABASE_STATE"
    if [[ "$DATABASE_STATE" == t ]]; then
      CURRENT_VERSION=$(PGPASSWORD="${DB_PASSWORD}" psql -qtAX -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}" -d "${DB_NAME}" -c "SELECT version FROM schema_migrations")
      echo "Current schema version: $CURRENT_VERSION"
      LAST_SUCCESSFUL_MIGRATION=$(ls -lr migrations/${MIGRATION_PATH} | grep -i -A 1 ${CURRENT_VERSION} | tail -n 1 | tr -s ' ' | cut -d ' ' -f9 | cut -d '_' -f1)
      echo "Forcing DB schema version to last successful migration version - $LAST_SUCCESSFUL_MIGRATION"
      migrate -path migrations/${MIGRATION_PATH} -database "$CONNECTION_STRING" force ${LAST_SUCCESSFUL_MIGRATION}
    fi
  fi

  echo '# STARTING MIGRATION #'
  $CMD
else
  REVERT_TO=$(ls -lr migrations/${MIGRATION_PATH} | head -n 2 | tail -n 1 | tr -s ' ' | cut -d ' ' -f9 | cut -d '_' -f1)
  LAST_SUCCESSFUL_MIGRATION=$(currentVersion)
  echo "Last successful migration is $LAST_SUCCESSFUL_MIGRATION"
  if [[ $LAST_SUCCESSFUL_MIGRATION == $CLEAN_DB_MESSAGE ]]; then
    REVERT_TO=$CLEAN_DB_MESSAGE
  fi

  echo "Will perform down migration to version $REVERT_TO"
  if [[ "$REVERT_TO" != "$CLEAN_DB_MESSAGE" ]]; then
    ensureMigrationExists ${MIGRATION_STORAGE_PATH} $REVERT_TO
  fi

  if [[ $LAST_SUCCESSFUL_MIGRATION != "$CLEAN_DB_MESSAGE" ]]; then
    LAST_SUCCESSFUL_MIGRATION=$(echo $LAST_SUCCESSFUL_MIGRATION | head -n1 | cut -d " " -f1)
    echo "Cleaning dirty flag - will force reset to last successful migration version \"$LAST_SUCCESSFUL_MIGRATION\""
    ensureMigrationExists ${MIGRATION_STORAGE_PATH} $LAST_SUCCESSFUL_MIGRATION
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
