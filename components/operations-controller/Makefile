APP_NAME = compass-operations-controller
APP_PATH = components/operations-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:v20210804-d3a9d497
VERIFY_IGNORE := /vendor\|/automock\|/api/v1alpha1/zz_generated.deepcopy.go
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
CRD_OPTIONS ?= "crd:trivialVersions=true"
CHART_PATH = $(realpath $(shell pwd)/../..)/chart/compass/charts/operations-controller
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

test-local: generate manifests
	go test ./...

build-local: generate
	env go build -o bin/manager cmd/main.go

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

copy-crds-to-chart:
	kustomize build config/crd > crds.yaml
	mv crds.yaml ${CHART_PATH}/templates/
