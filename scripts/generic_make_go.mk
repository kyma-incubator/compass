# Default configuration
ifneq ($(strip $(DOCKER_PUSH_REPOSITORY)),)
IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
else
IMG_NAME := $(APP_NAME)
endif

# Configuration for Kyma Environment Broker cleanup job image
ifneq ($(strip $(DOCKER_PUSH_REPOSITORY)),)
CLEANUP_IMG_NAME := $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_CLEANUP_NAME)
else
CLEANUP_IMG_NAME := $(APP_CLEANUP_NAME)
endif

ifneq ($(strip $(DOCKER_TAG)),)
TAG := $(DOCKER_TAG)
else
TAG := latest
endif
# BASE_PKG is a root packge of the component
BASE_PKG := github.com/kyma-incubator/compass
# IMG_GOPATH is a path to go path in the container
IMG_GOPATH := /workspace/go
# IMG_GOCACHE is a path to go cache in the container
IMG_GOCACHE := /root/.cache/go-build
# VERIFY_IGNORE is a grep pattern to exclude files and directories from verification
VERIFY_IGNORE := /vendor\|/automock

# Other variables
# LOCAL_DIR in a local path to scripts folder
LOCAL_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# COMPONENT_DIR is a local path to component
COMPONENT_DIR = $(shell pwd)
# COMPONENT_NAME is equivalent to the name of the component as defined in it's helm chart
COMPONENT_NAME = $(shell basename $(COMPONENT_DIR))
# WORKSPACE_LOCAL_DIR is a path to the scripts folder in the container
WORKSPACE_LOCAL_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/scripts
# WORKSPACE_COMPONENT_DIR is a path to component in hte container
WORKSPACE_COMPONENT_DIR = $(IMG_GOPATH)/src/$(BASE_PKG)/$(APP_PATH)
# FILES_TO_CHECK is a command used to determine which files should be verified
FILES_TO_CHECK = find . -type f -name "*.go" | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_CHECK is a command used to determine which directories should be verified
DIRS_TO_CHECK = go list ./... | grep -v "$(VERIFY_IGNORE)"
# DIRS_TO_IGNORE is a command used to determine which directories should not be verified
DIRS_TO_IGNORE = go list ./... | grep "$(VERIFY_IGNORE)"
# DEPLOYMENT_NAME matches the component's deployment name in the cluster
DEPLOYMENT_NAME="compass-"$(COMPONENT_NAME)
# NAMESPACE defines the namespace into which the component is deployed
NAMESPACE="compass-system"

# Base docker configuration
DOCKER_CREATE_OPTS := -v $(LOCAL_DIR):$(WORKSPACE_LOCAL_DIR):delegated --rm -w $(WORKSPACE_COMPONENT_DIR) $(BUILDPACK)

# Check if go is available
ifneq (,$(shell go version 2>/dev/null))
DOCKER_CREATE_OPTS := -v $(shell go env GOCACHE):$(IMG_GOCACHE):delegated -v $(shell go env GOPATH)/pkg/dep:$(IMG_GOPATH)/pkg/dep:delegated $(DOCKER_CREATE_OPTS)
endif

.DEFAULT_GOAL := verify

# Check if running with TTY
ifeq (1, $(shell [ -t 0 ] && echo 1))
DOCKER_INTERACTIVE := -i
DOCKER_CREATE_OPTS := -t $(DOCKER_CREATE_OPTS)
else
DOCKER_INTERACTIVE_START := --attach
endif

# Buildpack directives
define buildpack-mount
.PHONY: $(1)-local $(1)
$(1):
	@echo make $(1)
	@docker run $(DOCKER_INTERACTIVE) \
		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
		$(DOCKER_CREATE_OPTS) make $(1)-local
endef

define buildpack-cp-ro
.PHONY: $(1)-local $(1)
$(1):
	@echo make $(1)
	$$(eval container = $$(shell docker create $(DOCKER_CREATE_OPTS) make $(1)-local))
	@docker cp $(COMPONENT_DIR)/. $$(container):$(WORKSPACE_COMPONENT_DIR)/
	@docker start $(DOCKER_INTERACTIVE_START) $(DOCKER_INTERACTIVE) $$(container)
endef

.PHONY: verify format release check-gqlgen

# You may add additional targets/commands to be run on format and verify. Declare the target again in your makefile,
# using two double colons. For example to run errcheck on verify add this to your makefile:
#
#   verify:: errcheck
#
verify:: test check-imports check-fmt errcheck lint
format:: imports fmt

release: verify build-image

.PHONY: build-image
build-image: pull-licenses
	docker run --rm --privileged linuxkit/binfmt:v0.8 # https://stackoverflow.com/questions/70066249/docker-random-alpine-packages-fail-to-install
	docker buildx create --name multi-arch-builder --use
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMG_NAME):$(TAG) --push .
docker-create-opts:
	@echo $(DOCKER_CREATE_OPTS)

# Targets mounting sources to buildpack
MOUNT_TARGETS = build check-imports imports check-fmt fmt errcheck vet generate pull-licenses gqlgen lint
$(foreach t,$(MOUNT_TARGETS),$(eval $(call buildpack-mount,$(t))))

# Builds new Docker image into k3d's Docker Registry
build-for-k3d: pull-licenses-local
	docker build -t k3d-kyma-registry:5001/$(IMG_NAME):$(TAG) .
	docker push k3d-kyma-registry:5001/$(IMG_NAME):$(TAG)

build-local:
	env CGO_ENABLED=0 go build -o $(APP_NAME) ./$(ENTRYPOINT)
	rm $(APP_NAME)

check-imports-local:
	@if [ -n "$$(goimports -l $$($(FILES_TO_CHECK)))" ]; then \
		echo "✗ some files are not properly formatted or contain not formatted imports. To repair run make imports"; \
		goimports -l $$($(FILES_TO_CHECK)); \
		exit 1; \
	fi;

imports-local:
	goimports -w -l $$($(FILES_TO_CHECK))

check-fmt-local:
	@if [ -n "$$(gofmt -l $$($(FILES_TO_CHECK)))" ]; then \
		gofmt -l $$($(FILES_TO_CHECK)); \
		echo "✗ some files are not properly formatted. To repair run make fmt"; \
		exit 1; \
	fi;

fmt-local:
	go fmt $$($(DIRS_TO_CHECK))

format-local: imports-local fmt-local

verify-local: test-local check-imports-local check-fmt-local errcheck-local lint-local

errcheck-local:
	errcheck -blank -asserts -ignorepkg '$$($(DIRS_TO_CHECK) | tr '\n' ',')' -ignoregenerated ./...

vet-local:
	go vet $$($(DIRS_TO_CHECK))

lint-local:
	golangci-lint run

generate-local:
	go generate ./...

gqlgen-local:
	./gqlgen.sh

check-gqlgen:
	@echo make check-gqlgen
	@if [ -n "$$(git status -s pkg/graphql | grep -v '_test.go' | grep -v 'graphqlizer.go')" ]; then \
		echo -e "${RED}✗ gqlgen.sh modified some files, schema and code are out-of-sync${NC}"; \
		git status -s pkg/graphql | grep -v "_test.go" | grep -v 'graphqlizer.go'; \
		exit 1; \
	fi;

pull-licenses-local:
ifdef LICENSE_PULLER_PATH
	bash $(LICENSE_PULLER_PATH)
else
	mkdir -p licenses
endif

# Targets copying sources to buildpack
COPY_TARGETS = test
$(foreach t,$(COPY_TARGETS),$(eval $(call buildpack-cp-ro,$(t))))

test-local-no-coverage:
	go test ./...

test-local:
	@go test ./... -coverprofile cover.out.tmp
	@cat cover.out.tmp | grep -v "_gen.go" | grep -v "hack" > cover.out
	@echo -n 'Code Coverage: '
	@go tool cover -func cover.out | grep total | tr -s "[:blank:]" | tr "[:blank:]" " " | cut -d " " -f 3
	@rm cover.out.tmp cover.out

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(COMPONENT_DIR)/Makefile : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

.PHONY: exec
exec:
	@docker run $(DOCKER_INTERACTIVE) \
    		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
    		$(DOCKER_CREATE_OPTS) bash

# Sets locally built image for a given component in k3d cluster
deploy-on-k3d: build-for-k3d
	kubectl config use-context k3d-kyma
	kubectl patch -n $(NAMESPACE) deployment/$(DEPLOYMENT_NAME) -p '{"spec":{"template":{"spec":{"containers":[{"name":"'$(COMPONENT_NAME)'","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/$(DEPLOYMENT_NAME) $(COMPONENT_NAME)=k3d-kyma-registry:5001/$(DEPLOYMENT_NAME):$(TAG)
	kubectl rollout restart -n $(NAMESPACE) deployment/$(DEPLOYMENT_NAME)
