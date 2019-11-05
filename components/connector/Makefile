APP_NAME = compass-connector
APP_PATH = components/connector
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/mocks

verify:: gqlgen check-gqlgen
