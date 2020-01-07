APP_NAME = kyma-environment-broker
APP_PATH = components/kyma-environment-broker
ENTRYPOINT = cmd/broker/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
DOCKER_SOCKET = /var/run/docker.sock
TESTING_DB_NETWORK = test_network

include $(SCRIPTS_DIR)/generic_make_go.mk

.DEFAULT_GOAL := custom-verify

.PHONY: testing-with-database-network check-imports check-fmt release

custom-verify: testing-with-database-network check-imports check-fmt

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
		$(DOCKER_CREATE_OPTS) go test ./...
	@docker network rm $(TESTING_DB_NETWORK) || true

clean-up:
	@docker network rm $(TESTING_DB_NETWORK) || true
