APP_NAME = istio-tests
APP_PATH = tests/istio
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

test-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh istio

e2e-test-clean:
	kubectl config use-context k3d-kyma
	kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

sandbox-test:
	@../sandbox.sh istio

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-istio-local-0 -c istio-tests -- ./istio.test test -run=$(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o istio.test
	kubectl cp ./istio.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-istio-local-0:/ -c istio-tests
	rm ./istio.test

sandbox-test-clean:
	kubectl config use-context k3d-kyma
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-istio-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true