APP_NAME = system-broker-tests
APP_PATH = tests/system-broker
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20201208-f3aa8f74
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

test-local: ;

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@sed "s/PLACEHOLDER/system-broker/" ../test-suite.yaml | kubectl -n kyma-system apply -f -

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests

