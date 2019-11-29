APP_NAME = compass-provisioner
APP_PATH = components/provisioner
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts

include $(SCRIPTS_DIR)/generic_make_go.mk

verify:: gqlgen check-gqlgen
