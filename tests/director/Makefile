APP_NAME = compass-director-tests
APP_PATH = tests/director-tests
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210125-6234473e
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
DIRECTOR_GRAPHQL_API = "http://compass-dev-director:3000"
export DIRECTOR_GRAPHQL_API
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk

# We have to override test-local, because we need to run director with database as docker containers and connected with custom network
# and the container itself has to be connected to the network
test-local:
	@echo ${SKIP_STEP_MESSAGE}

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

e2e-test:
	@../../installation/scripts/testing.sh director

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests

generate-examples:
	@./generate_examples.sh