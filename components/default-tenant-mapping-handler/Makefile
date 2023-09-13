APP_NAME = compass-default-tenant-mapping-handler
APP_PATH = components/default-tenant-mapping-handler
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
NAMESPACE="compass-system"
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata


errcheck-local:
	errcheck -blank -asserts -exclude ./excluded.txt -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

#lint-local:
#	@echo ${SKIP_STEP_MESSAGE}

deploy-on-k3d: build-for-k3d
	kubectl config use-context k3d-kyma
	kubectl patch -n $(NAMESPACE) deployment/compass-default-tenant-mapping-handler -p '{"spec":{"template":{"spec":{"containers":[{"name":"default-tenant-mapping-handler","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-default-tenant-mapping-handler default-tenant-mapping-handler=k3d-kyma-registry:5001/compass-default-tenant-mapping-handler:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-default-tenant-mapping-handler