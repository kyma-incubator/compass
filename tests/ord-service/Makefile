APP_NAME = ord-service-tests
COMPONENT = ord-service
APP_PATH = tests/ord-service
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.17.0
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh ord-service

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh ord-service

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-ord-service-local-0 -c ord-service-tests -- ./ord-service.test -test.run $(testName) -test.v

bench-run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-ord-service-local-0 -c ord-service-tests -- ./ord-service.bench -test.bench $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux go test -c ./tests -o ord-service.test
	kubectl cp ./ord-service.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-ord-service-local-0:/ -c ord-service-tests
	rm ./ord-service.test

sandbox-deploy-bench-test:
	env GOOS=linux go test -c ./bench -o ord-service.bench
	kubectl cp ./ord-service.bench kyma-system/oct-tp-compass-e2e-tests-compass-e2e-ord-service-local-0:/ -c ord-service-tests
	rm ./ord-service.bench

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-ord-service-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true