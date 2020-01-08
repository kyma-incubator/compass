APP_NAME = kyma-environment-broker
APP_PATH = components/kyma-environment-broker
ENTRYPOINT = cmd/broker/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk
