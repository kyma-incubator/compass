APP_NAME = compass-system-broker
APP_PATH = components/system-broker
COMPONENT_BUILD_TIMEOUT = 10m
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
ADDITIONAL_COMPONENTS = director connector
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/*fakes

verify:: gqlgen check-gqlgen

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

# System Broker component tests depend on the Director's schema.graphql file, so we had to redefine the
# test target to also mount the Director folder.
test:
	@echo make test
	$(eval container = $(shell docker create $(DOCKER_CREATE_OPTS) make test-local))
	@docker cp $(COMPONENT_DIR)/. $(container):$(WORKSPACE_COMPONENT_DIR)/
	@docker cp $(COMPONENT_DIR)/../director $(container):$(WORKSPACE_COMPONENT_DIR)/../director
	@docker start $(DOCKER_INTERACTIVE_START) $(DOCKER_INTERACTIVE) $(container)

test-local:
	go test ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

gqlgen-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	env go build -o bin/system-broker ./cmd/main.go
