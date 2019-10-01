APP_NAME = compass-director
APP_PATH = components/director
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

.PHONY: ci-pr ci-master ci-release
include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

ci-pr: resolve build-local test-local build-image push-image
ci-master: resolve build-local test-local build-image push-image
ci-release: resolve build-local test-local build-image push-image
