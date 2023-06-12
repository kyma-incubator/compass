APP_NAME = compass-kyma-adapter
APP_PATH = components/kyma-adapter
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
NAMESPACE="compass-system"
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata


errcheck-local:
	errcheck -blank -asserts -exclude ./excluded.txt -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

#lint-local:
#	@echo ${SKIP_STEP_MESSAGE}
