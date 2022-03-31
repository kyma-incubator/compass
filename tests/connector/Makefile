APP_NAME = connector-tests
APP_PATH = tests/connector
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: vet

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh connector

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh connector

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-connector-local-0 -c connector-tests -- ./connector.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux go test -c ./tests -o connector.test
	kubectl cp ./connector.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-connector-local-0:/ -c connector-tests
	rm ./connector.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-connector-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true