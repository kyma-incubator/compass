APP_NAME = connectivity-adapter-tests
APP_PATH = tests/connectivity-adapter
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh connectivity-adapter

e2e-test-clean:
	kubectl config use-context k3d-kyma
	kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh connectivity-adapter

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-connectivity-adapter-local-0 -c connectivity-adapter-tests -- ./connectivity-adapter.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o connectivity-adapter.test
	kubectl cp ./connectivity-adapter.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-connectivity-adapter-local-0:/ -c connectivity-adapter-tests
	rm ./connectivity-adapter.test

sandbox-test-clean:
	kubectl config use-context k3d-kyma
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-connectivity-adapter-local  || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true