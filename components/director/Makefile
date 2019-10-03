APP_NAME = compass-director
APP_PATH = components/director
ENTRYPOINT = cmd/main.go
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190913-65b55d1
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts


include $(SCRIPTS_DIR)/go-dep.mk

VERIFY_IGNORE := /vendor\|/automock\|/testdata\|/pkg

gqlgen:
	exit $(./gqlgen.sh)

gqlgen-check:
	@if [ -n $(git status -s pkg/graphql) ]; then\
		echo $(git status -s ./pkg/graphql);\
		echo -e "${RED}✗ gqlgen.sh modified some files, schema and code are out-of-sync${NC}";\
		exit 1;\
	fi

verify:: gqlgen
