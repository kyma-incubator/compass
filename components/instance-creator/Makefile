APP_NAME = compass-instance-creator
APP_PATH = components/instance-creator
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
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
	kubectl patch -n $(NAMESPACE) deployment/compass-instance-creator -p '{"spec":{"template":{"spec":{"containers":[{"name":"instance-creator","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-instance-creator instance-creator=k3d-kyma-registry:5001/compass-instance-creator:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-instance-creator