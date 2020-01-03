APP_NAME = compass-director
APP_PATH = components/director
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: gqlgen check-gqlgen

errcheck-local:
	 GO111MODULE="on" errcheck -blank -asserts -ignoregenerated ./...

test-local:
	GO111MODULE=on go test ./...

resolve-local: ;

ensure-local: ;

dep-status-local: ;
