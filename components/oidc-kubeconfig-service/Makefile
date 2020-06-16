APP_NAME = kubeconfig-service
APP_PATH = components/kubeconfig-service
ENTRYPOINT = cmd/generator/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200423-1d9d6590
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

export GO111MODULE=on
export CGO_ENABLED=0
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}
