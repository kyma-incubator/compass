APP_NAME = compass-schema-migrator
APP_PATH = components/schema-migrator
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

.PHONY: validate

validate:
	./validate.sh

verify:: validate

.PHONY: ci-pr
#Temporary
ci-pr: release
