
APP_NAME = compass-ias-adapter
APP_PATH = components/ias-adapter
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
NAMESPACE="compass-system"
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

run:
	export ENV="dev" && go run cmd/adapter/main.go

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
	env go build -o bin/ias-adapter ./cmd/adapter/main.go

install-tools:
	go mod download
	@echo Installing tools from tools.go
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

deploy-on-k3d: build-for-k3d
	kubectl config use-context k3d-kyma
	kubectl patch -n $(NAMESPACE) deployment/compass-ias-adapter -p '{"spec":{"template":{"spec":{"containers":[{"name":"ias-adapter","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-ias-adapter ias-adapter=k3d-kyma-registry:5001/compass-ias-adapter:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-ias-adapter