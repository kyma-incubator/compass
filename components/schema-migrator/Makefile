APP_NAME = compass-schema-migrator
APP_PATH = components/schema-migrator
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export SKIP_DEPLOY_MESSAGE = "Building minikube image and redeployment of Schema Migrator is not allowed"
export SKIP_TEST_MESSAGE = "Test goal for Schema Migrator is empty"
export SKIP_BUILD_MESSAGE = "Build goal for Schema Migrator is empty"
export SKIP_CHECK_IMPORTS_MESSAGE = "Check imports goal for Schema Migrator is empty"
export SKIP_CHECK_FMT_MESSAGE = "Check fmt goal for Schema Migrator is empty"

include $(SCRIPTS_DIR)/generic_make_go.mk

.PHONY: validate

validate:
	@if [ "$(COMPASS_DUMP_DB)" == "true" ]; then\
		./validate.sh --dump-db;\
    else\
		./validate.sh;\
    fi

verify:: validate clean-dump

test-local:
	@echo ${SKIP_TEST_MESSAGE}

build-local:
	@echo ${SKIP_BUILD_MESSAGE}

check-imports:
	@echo ${SKIP_CHECK_IMPORTS_MESSAGE}

check-fmt:
	@echo ${SKIP_CHECK_FMT_MESSAGE}

deploy-on-minikube:
	@echo ${SKIP_DEPLOY_MESSAGE}

clean-dump:
	@rm -f ./seeds/dump.sql