APP_NAME = compass-provisioner
APP_PATH = components/provisioner
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
DOCKER_SOCKET = /var/run/docker.sock
TESTING_DB_NETWORK = test_network
PIPELINE_BUILD = 1
export PIPELINE_BUILD

include $(SCRIPTS_DIR)/generic_make_go.mk

.PHONY: gqlgen check-gqlgen testing-with-database-network clean-up mod-verify go-mod-check

verify:: gqlgen check-gqlgen testing-with-database-network mod-verify go-mod-check

resolve-local:
	GO111MODULE=on go mod vendor -v

ensure-local:
	@echo "Go modules present in component - omitting."

dep-status:
	@echo "Go modules present in component - omitting."

dep-status-local:
	@echo "Go modules present in component - omitting."

mod-verify-local:
	GO111MODULE=on go mod verify

go-mod-check-local:
	@echo make go-mod-check
	go mod tidy
	@if [ -z "$$(git status -s go.*)" ]; then \
		echo -e "${RED}✗ go mod tidy modified go.mod or go.sum files${NC}"; \
		git status -s git status -s go.*; \
		exit 1; \
	fi;


# We have to override test-local and errcheck, because we need to run provisioner with database
#as docker container connected with custom network and the buildpack container itsefl has to be connected to the network

test-local: ;

# TODO: there is no errcheck in go1.13 buildpack, consider creating buildpack-toolbox with go1.13 version
errcheck-local: ;
# errcheck-local:
# 	@docker run $(DOCKER_INTERACTIVE) \
# 		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
# 		$(DOCKER_CREATE_OPTS) errcheck -blank -asserts -ignorepkg '$$($(DIRS_TO_CHECK) | tr '\n' ',')' -ignoregenerated ./...


testing-with-database-network:
	@echo testing-with-database-network
	@docker network inspect $(TESTING_DB_NETWORK) >/dev/null 2>&1 || \
	docker network create --driver bridge $(TESTING_DB_NETWORK)
	@docker run $(DOCKER_INTERACTIVE) \
		-v $(DOCKER_SOCKET):$(DOCKER_SOCKET) \
		-v $(COMPONENT_DIR)/../../:$(WORKSPACE_COMPONENT_DIR)/../../ \
		--network=$(TESTING_DB_NETWORK) \
		-v $(COMPONENT_DIR):$(WORKSPACE_COMPONENT_DIR):delegated \
		--env PIPELINE_BUILD=1 --env GO111MODULE=on \
		$(DOCKER_CREATE_OPTS) go test ./...
	@docker network rm $(TESTING_DB_NETWORK)

clean-up:
	@docker network rm $(TESTING_DB_NETWORK)
