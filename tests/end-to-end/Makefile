APP_NAME = compass-end-to-end-test
APP_PATH = tests/end-to-end
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
DIRECTOR_GRAPHQL_API = "http://compass-dev-director:3000"
export DIRECTOR_GRAPHQL_API

include $(SCRIPTS_DIR)/generic_make_go.mk

.PHONY: generate-director-examples clean-up

VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: generate-director-examples

$(eval $(call test-local))

# We have to override test-local, because we need to run director with database as docker containers and connected with custom network
# and the container itsefl has to be connected to the network
test-local: ;

generate-director-examples: clean-up
	@echo make generate-director-examples
	@docker network create --driver bridge gen-examples-network
	@docker run $(DOCKER_INTERACTIVE) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(COMPONENT_DIR)/../../:$(WORKSPACE_COMPONENT_DIR)/../../ \
		-v $(COMPONENT_DIR)/../../components/director:$(WORKSPACE_COMPONENT_DIR)/../../components/director:delegated \
		-e HOST_ROOT_PATH=$(COMPONENT_DIR)/../.. \
		--network=gen-examples-network \
		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
		$(DOCKER_CREATE_OPTS) 	./gen-examples.sh
	@docker network rm gen-examples-network > /dev/null 2>&1
	@if [ -n "$$(git status -s ../../components/director/examples )" ]; then \
		echo -e "✗ Code and examples are out-of-sync" \
		git status -s ; \
		exit 1; \
	fi;

clean-up:
	@docker network rm gen-examples-network > /dev/null 2>&1 || true
