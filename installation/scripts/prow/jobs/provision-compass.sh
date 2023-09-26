#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with compass. It takes an optional positional parameter using --image flag
# Use this flag to specify the custom image for provisioning vms. If no flag is provided, the latest custom image is used.

set -o errexit

readonly COMPASS_SOURCE_DIR="/home/prow/go/src/github.com/kyma-incubator/compass"
readonly TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
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

# Download yq
YQ_VERSION="v4.25.1"
log::info "Downloading yq version: $YQ_VERSION"
curl -fsSL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -o yq
chmod +x yq
cp yq "$HOME/bin/yq" && cp yq "/usr/local/bin/yq"
log::info "Successfully installed yq version: $YQ_VERSION"

# Install Kyma to be later used in run.sh
# KYMA_CLI_VERSION=$(cat ${COMPASS_SOURCE_DIR}/installation/resources/KYMA_VERSION)
# TODO: Kyma 2.5.2 release exists, but Kyma CLI with the same version does not. That's why we're using 2.6.2
KYMA_CLI_VERSION="2.6.2"
log::info "Installing Kyma CLI version: $KYMA_CLI_VERSION"

curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/${KYMA_CLI_VERSION}/kyma_Linux_x86_64.tar.gz" \
&& mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && mv kyma-release/kyma /usr/local/bin \
&& rm -rf kyma-release kyma.tar.gz

log::info "Successfully installed Kyma CLI version: $KYMA_CLI_VERSION"

# Install openssl which is later used in run.sh
log::info "Installing openssl..."
apk add openssl
log::info "Successfully installed openssl"

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --image)
            IMAGE="$2"
            testCustomImage "${IMAGE}"
            shift
            shift
            ;;
        --dump-db)
            DUMP_DB="--dump-db"
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

log::info "Triggering the compass installation"
${COMPASS_SOURCE_DIR}/installation/scripts/prow/provision.sh ${DUMP_DB}
log::info "Compass provisioning done"

log::info "Triggering the tests"
${COMPASS_SOURCE_DIR}/installation/scripts/prow/execute-tests.sh ${DUMP_DB}
log::info "Test execution completed"
