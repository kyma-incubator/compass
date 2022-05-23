APP_NAME = tenant-fetcher-job-tests
APP_PATH = tests/tenant-fetcher-job
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for E2E tests project"

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh tenant-fetcher-job

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh tenant-fetcher-job

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-tenant-fetcher-job-local-0 -c tenant-fetcher-job-tests -- ./tenant-fetcher-job.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux go test -c ./tests -o tenant-fetcher-job.test
	kubectl cp ./tenant-fetcher-job.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-tenant-fetcher-job-local-0:/ -c tenant-fetcher-job-tests
	rm ./tenant-fetcher-job.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-tenant-fetcher-job-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true