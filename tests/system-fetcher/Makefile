APP_NAME = system-fetcher-tests
APP_PATH = tests/system-fetcher
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh system-fetcher

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh system-fetcher

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-system-fetcher-local-0 -c system-fetcher-tests -- ./system-fetcher.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o system-fetcher.test
	kubectl cp ./system-fetcher.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-system-fetcher-local-0:/ -c system-fetcher-tests
	rm ./system-fetcher.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-system-fetcher-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true