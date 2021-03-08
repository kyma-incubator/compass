APP_NAME = tenant-fetcher-tests
APP_PATH = tests/tenant-fetcher
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

test-local: ;

test:
	@sed "s/PLACEHOLDER/tenant-fetcher/" ../test-suite.yaml | kubectl -n kyma-system apply -f -

test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests