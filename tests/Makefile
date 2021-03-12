APP_NAME = compass-tests
APP_PATH = tests
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	sed '10,14d' test-suite.yaml | kubectl -n kyma-system apply -f -

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests

deploy-on-minikube:
	@eval $$(minikube docker-env) && docker build -t eu.gcr.io/kyma-project/incubator/compass-e2e-tests .