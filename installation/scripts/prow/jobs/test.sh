readonly TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

KYMA_CLI_VERSION="2.3.0"
echo "Installing KYMA"

PREV_WD=$(pwd)
git clone https://github.com/kyma-project/cli.git && cd cli && git checkout $KYMA_CLI_VERSION
make build-linux && cd ./bin && mv ./kyma-linux ./kyma
ls
chmod +x ./kyma
KYMA_DIR="${pwd}/kyma"
export PATH="$PATH:${KYMA_DIR}"

echo "Kyma CLI installed with version: $KYMA_CLI_VERSION"
echo "Installing Compass"

kyma version

../../installation/cmd/run.sh

echo "Compass installed"
