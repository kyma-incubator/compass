APP_NAME = pairing-adapter-tests
COMPONENT = pairing-adapter
APP_PATH = tests/pairing-adapter
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for e2e tests project"

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh pairing-adapter

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh pairing-adapter

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-pairing-adapter-local-0 -c pairing-adapter-tests -- ./pairing-adapter.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o pairing-adapter.test
	kubectl cp ./pairing-adapter.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-pairing-adapter-local-0:/ -c pairing-adapter-tests
	rm ./pairing-adapter.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-pairing-adapter-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true