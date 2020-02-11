APP_NAME = kyma-environment-broker
APP_PATH = components/kyma-environment-broker
ENTRYPOINT = cmd/broker/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
DOCKER_SOCKET = /var/run/docker.sock
TESTING_DB_NETWORK = test_network

include $(SCRIPTS_DIR)/generic_make_go.mk

.DEFAULT_GOAL := custom-verify

custom-verify: testing-with-database-network test-integration check-imports check-fmt

verify:: custom-verify build-image push-image

# We have to override test-local and errcheck, because we need to run provisioner with database
#as docker container connected with custom network and the buildpack container itsefl has to be connected to the network

test-local: ;
errcheck-local: ;

# TODO: there is no errcheck in go1.13 buildpack, consider creating buildpack-toolbox with go1.13 version
# errcheck-local:
# 	@docker run $(DOCKER_INTERACTIVE) \
# 		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
# 		$(DOCKER_CREATE_OPTS) errcheck -blank -asserts -ignorepkg '$$($(DIRS_TO_CHECK) | tr '\n' ',')' -ignoregenerated ./...

test-integration-local:
	go test ./... -tags=integration

test-integration:
	@echo make test-integration-local
	@docker run $(DOCKER_INTERACTIVE) \
		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
		$(DOCKER_CREATE_OPTS) make test-integration-local

testing-with-database-network:
	@echo testing-with-database-network
	@docker network inspect $(TESTING_DB_NETWORK) >/dev/null 2>&1 || \
	docker network create --driver bridge $(TESTING_DB_NETWORK)
	@docker run $(DOCKER_INTERACTIVE) \
		-v $(DOCKER_SOCKET):$(DOCKER_SOCKET) \
		-v $(COMPONENT_DIR)/../../:$(WORKSPACE_COMPONENT_DIR)/../../ \
		--network=$(TESTING_DB_NETWORK) \
		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
		--env PIPELINE_BUILD=1 \
		$(DOCKER_CREATE_OPTS) go test  -tags=database-integration ./...
	@docker network rm $(TESTING_DB_NETWORK) || true

clean-up:
	@docker network rm $(TESTING_DB_NETWORK) || true
