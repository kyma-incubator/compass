APP_NAME = compass-director-tests
APP_PATH = tests/director-tests
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
SCRIPTS_DIR = $(realpath $(shell pwd)/../../../../)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../../../../)/installation/scripts
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

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh director-runtime

e2e-test-clean:
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

generate-examples:
	@../generate_examples.sh

sandbox-test:
	@../../../sandbox.sh director-runtime

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-director-runtime-local-0 -c director-runtime-tests -- ./director-runtime.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c . -o director-runtime.test
	kubectl cp ./director-runtime.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-director-runtime-local-0:/ -c director-runtime-tests
	rm ./director-runtime.test

sandbox-test-clean:
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-director-runtime-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true