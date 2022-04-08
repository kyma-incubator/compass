APP_NAME = compass-director
APP_PATH = components/director
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.17.0
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: gqlgen check-gqlgen

errcheck-local:
	errcheck -blank -asserts -exclude ./excluded.txt -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	env go build -o bin/director ./cmd/director/main.go
	env go build -o bin/tenantfetcher-job ./cmd/tenantfetcher-job/main.go
	env go build -o bin/tenantfetcher-svc ./cmd/tenantfetcher-svc/main.go

install-tools:
	go mod download
	@echo Installing tools from tools.go
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
