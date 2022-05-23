APP_NAME = ord-aggregator-tests
APP_PATH = tests/ord-aggregator
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
export GO111MODULE = on

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh ord-aggregator

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh ord-aggregator

run:
	 @kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-ord-aggregator-local-0 -c ord-aggregator-tests -- ./ord-aggregator.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux go test -c ./tests -o ord-aggregator.test
	kubectl cp ./ord-aggregator.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-ord-aggregator-local-0:/ -c ord-aggregator-tests
	rm ./ord-aggregator.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-ord-aggregator-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

