APP_NAME = compass-external-services-mock
APP_PATH = components/external-services-mock
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

make errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

test-local:
	go test ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	env go build -o bin/external-services-mock ./cmd/main.go

deploy-on-k3d: build-for-k3d
	kubectl config use-context k3d-kyma
	kubectl patch -n $(NAMESPACE) deployment/compass-external-services-mock -p '{"spec":{"template":{"spec":{"containers":[{"name":"external-services-mock","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-external-services-mock external-services-mock=k3d-kyma-registry:5001/compass-external-services-mock:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-external-services-mock
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-gateway
