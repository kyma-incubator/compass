#!/usr/bin/env bash

set -o errexit

readonly TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"

if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(pull-.*)"
    export JOBGUARD_TIMEOUT="60m"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

log::info "Installing gcloud CLI"
apk add --no-cache python3 py3-crcmod py3-openssl
export PATH="/google-cloud-sdk/bin:${PATH}"
GCLOUD_CLI_VERSION="437.0.1"
curl -fLSs -o gc-sdk.tar.gz "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_CLI_VERSION}-linux-x86_64.tar.gz"
tar xzf gc-sdk.tar.gz -C /
rm gc-sdk.tar.gz
gcloud components install alpha beta kubectl docker-credential-gcr gke-gcloud-auth-plugin
gcloud config set core/disable_usage_reporting true
gcloud config set component_manager/disable_update_check true
gcloud config set metrics/environment github_docker_image
gcloud --version

log::info "Authenticate to GCP through gcloud"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Authenticate"
gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Start Docker"
docker::start

chmod -R 0777 /home/prow/go/src/github.com/kyma-incubator/compass/.git

log::info "Read parameters"

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --component)
            COMPONENT="$2"
            shift
            shift
            ;;
        --command)
            COMMAND="$2"
            shift
            shift
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

if [[ -z "${COMPONENT}" ]]; then
log::error "There are no --component specified, the script will exit ..." && exit 1
fi   

if [[ -z "${COMMAND}" ]]; then
log::error "There are no --command specified, the script will exit ..." && exit 1
fi   

log::info "Triggering the validation with component: ${COMPONENT} and command: ${COMMAND}"

cd "/home/prow/go/src/github.com/kyma-incubator/compass/components/${COMPONENT}" && ${COMMAND}

log::info "Validation finished"