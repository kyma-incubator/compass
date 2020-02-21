APP_NAME = connectivity-adapter
APP_PATH = components/connectivity-adapter
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

errcheck-local:
	 GO111MODULE="on" errcheck -blank -asserts -ignoregenerated ./...

test-local:
	GO111MODULE=on go test ./...

resolve-local: ;

ensure-local: ;

dep-status-local: ;
