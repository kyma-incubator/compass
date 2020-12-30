APP_NAME = compass-ord-service
APP_PATH = components/ord-service
BUILDPACK = buildpack-java:local
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export COMPONENT_DIR = $(realpath $(shell pwd))
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk

release:
	@echo ${SKIP_STEP_MESSAGE}

errcheck-local:
	@echo ${SKIP_STEP_MESSAGE}

test-local:
	./scripts/install_dependencies.sh
	./scripts/build.sh

resolve-local:
	./scripts/install_dependencies.sh
	./scripts/build.sh -DskipTests

ensure-local:
	./scripts/install_dependencies.sh
	./scripts/build.sh -DskipTests

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	./scripts/build.sh -DskipTests