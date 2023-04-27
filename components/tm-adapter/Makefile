APP_NAME = compass-tm-adapter
APP_PATH = components/tm-adapter
#BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20230323-a1ac96db # 1.20
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
NAMESPACE="compass-system"
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: gqlgen check-gqlgen

errcheck-local:
	errcheck -blank -asserts -exclude ./excluded.txt -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	env go build -o bin/tm-adapter ./cmd/main.go

install-tools:
	go mod download
	@echo Installing tools from tools.go
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

deploy-on-k3d: build-for-k3d
	kubectl config use-context k3d-kyma
	kubectl patch -n $(NAMESPACE) deployment/compass-tm-adapter -p '{"spec":{"template":{"spec":{"containers":[{"name":"tm-adapter","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-tm-adapter tm-adapter=k3d-kyma-registry:5001/compass-tm-adapter:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-tm-adapter
