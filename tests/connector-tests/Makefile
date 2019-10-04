APP_NAME = connector-tests
APP_PATH = tests/connector-tests
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: vet

#We have to override test-local, because
test-local: ;

.PHONY: ci-pr
#Temporary
ci-pr: release
