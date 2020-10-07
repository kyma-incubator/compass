APP_NAME = compass-system-broker
APP_PATH = components/system-broker
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200423-1d9d6590
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/*fakes

verify:: gqlgen check-gqlgen

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

test-local:
	go test ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	env go build -o bin/system-broker ./cmd/main.go
