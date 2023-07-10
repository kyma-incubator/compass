readonly TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

KYMA_CLI_VERSION="2.3.0"
log::info "Installing Kyma CLI version: $KYMA_CLI_VERSION"

PREV_WD=$(pwd)
git clone https://github.com/kyma-project/cli.git && cd cli && git checkout $KYMA_CLI_VERSION
make build-linux && cd ./bin && mv ./kyma-linux ./kyma
chmod +x kyma

log::info "Kyma CLI installed with version: $KYMA_CLI_VERSION"
log::info "Installing Compass"

ls -R
cd ../../ && ls

../../installation/cmd/run.sh

log::info "Compass installed"
