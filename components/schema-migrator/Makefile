APP_NAME = compass-schema-migrator
APP_PATH = components/schema-migrator
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export SKIP_DEPLOY_MESSAGE = "Building minikube image and redeployment of Schema Migrator is not allowed"

include $(SCRIPTS_DIR)/generic_make_go.mk

.PHONY: validate

validate:
	./validate.sh

verify:: validate

build-to-minikube:
	@echo ${SKIP_DEPLOY_MESSAGE}

deploy-on-minikube:
	@echo ${SKIP_DEPLOY_MESSAGE}
