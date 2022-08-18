APP_NAME = compass-tests
APP_NAME = compass-e2e-tests
APP_PATH = tests
LOCAL_PREFIX = k3d-kyma-registry:5001
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/..)/installation/scripts
CHARTS_DIR = $(realpath $(shell pwd)/..)/chart
IMAGE_VERSION = $(shell grep e2e_tests --after-context=2 ${CHARTS_DIR}/compass/values.yaml | grep version | cut -d ':' -f 2 | xargs)
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

.DEFAULT_GOAL := errcheck-local

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

test:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

deploy-tests-on-k3d:
	docker build -t ${LOCAL_PREFIX}/compass-e2e-tests:${IMAGE_VERSION} .
	docker push ${LOCAL_PREFIX}/compass-e2e-tests:${IMAGE_VERSION}
	kubectl config use-context k3d-kyma
	$(INSTALLATION_SCRIPTS_DIR)/patch-test-definitions.sh ${IMAGE_VERSION}
