APP_NAME = compass-gateway
APP_PATH = components/gateway
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on

include $(SCRIPTS_DIR)/generic_make_go.mk
