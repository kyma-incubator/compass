APP_NAME = system-broker-tests
APP_PATH = tests/system-broker
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh system-broker

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh system-broker

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-system-broker-local-0 -c system-broker-tests -- ./system-broker.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o system-broker.test
	kubectl cp ./system-broker.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-system-broker-local-0:/ -c system-broker-tests
	rm ./system-broker.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-system-broker-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

