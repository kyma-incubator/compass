APP_NAME = compass-ns-adapter-tests
APP_PATH = tests/ns-adapter-tests
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk


test-local:
	@echo ${SKIP_STEP_MESSAGE}

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh ns-adapter

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

generate-examples:
	@./generate_examples.sh

sandbox-test:
	@../sandbox.sh ns-adapter

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-ns-adapter-local-0 -c ns-adapter-tests -- ./ns-adapter.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o ns-adapter.test
	kubectl cp ./ns-adapter.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-ns-adapter-local-0:/ -c ns-adapter-tests
	rm ./ns-adapter.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-ns-adapter-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true