APP_NAME = operations-controller-tests
APP_PATH = tests/operations-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20210804-d3a9d497
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh operations-controller

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh operations-controller

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-operations-controller-local-0 -c operations-controller-tests -- ./operations-controller.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=arm go test -c ./tests -o operations-controller.test
	kubectl cp ./operations-controller.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-operations-controller-local-0:/ -c operations-controller-tests
	rm ./operations-controller.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-operations-controller-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true