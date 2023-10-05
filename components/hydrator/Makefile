APP_NAME = compass-hydrator
APP_PATH = components/hydrator
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
ENTRYPOINT = cmd/main.go
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/mocks

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}
