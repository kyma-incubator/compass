APP_NAME = compass-connectivity-adapter
APP_PATH = components/connectivity-adapter
COMPONENT_BUILD_TIMEOUT = 10m
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

errcheck-local:
	 GO111MODULE="on" errcheck -blank -asserts -ignoregenerated ./...

test-local:
	GO111MODULE=on go test ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}
