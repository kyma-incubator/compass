APP_NAME = compass-fake-external-test-component
APP_PATH = components/fake-external-test-component
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

make errcheck-local:
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
	env go build -o bin/director ./cmd/director/main.go
	env go build -o bin/tenantfetcher ./cmd/tenantfetcher/main.go
