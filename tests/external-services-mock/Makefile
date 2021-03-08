APP_NAME = external-services-mock-tests
APP_PATH = tests/external-services-mock
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: vet

test-local: ;

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

test:
	@sed "s/PLACEHOLDER/external-services-mock/" ../test-suite.yaml | kubectl -n kyma-system apply -f -

test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests