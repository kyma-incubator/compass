APP_NAME = ord-service-tests
COMPONENT = ord-service
APP_PATH = tests/ord-service
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

test-local: ;

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

test:
	@sed "s/PLACEHOLDER/ord-service/" ../test-suite.yaml | kubectl -n kyma-system apply -f -

test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests

#sandbox-test:
#	@../sandbox.sh ord-service
#
#sandbox-test-clean:
#	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests
#	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-$(COMPONENT)-local