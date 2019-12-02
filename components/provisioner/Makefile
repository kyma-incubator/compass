APP_NAME = compass-provisioner
APP_PATH = components/provisioner
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

verify:: gqlgen check-gqlgen mod-verify

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

test-local:
	GO111MODULE=on go test ./...

# TODO - there is no errcheck in go1.13 buildpack, consider creating buildpack-toolbox with go1.13 version
errcheck-local: ;

go-mod-check-local:
	@echo make go-mod-check
	go mod tidy
	@if [ -z "$$(git status -s go.*)" ]; then \
		echo -e "${RED}✗ go mod tidy modified go.mod or go.sum files${NC}"; \
		git status -s git status -s go.*; \
		exit 1; \
	fi;

$(eval $(call buildpack-cp-ro,resolve))
$(eval $(call buildpack-mount,mod-verify))
$(eval $(call buildpack-mount,test))
$(eval $(call buildpack-mount,go-mod-check))
