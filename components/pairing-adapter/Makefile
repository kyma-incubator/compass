APP_NAME = pairing-adapter
APP_PATH = components/pairing-adapter
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

errcheck-local:
	 GO111MODULE="on" errcheck -blank -asserts -ignoregenerated ./...

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

test-local:
	GO111MODULE=on go test ./...

resolve-local: ;

ensure-local: ;

dep-status-local: ;
