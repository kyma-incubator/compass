APP_NAME = tenant-fetcher-tests
APP_PATH = tests/tenant-fetcher
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh tenant-fetcher

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh tenant-fetcher

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-tenant-fetcher-local-0 -c tenant-fetcher-tests -- ./tenant-fetcher.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o tenant-fetcher.test
	kubectl cp ./tenant-fetcher.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-tenant-fetcher-local-0:/ -c tenant-fetcher-tests
	rm ./tenant-fetcher.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-tenant-fetcher-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true