APP_NAME = compass-pairing-adapter
APP_PATH = components/pairing-adapter
ENTRYPOINT = cmd/main.go
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
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
