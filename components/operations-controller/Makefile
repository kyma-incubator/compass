APP_NAME = compass-operations-controller
APP_PATH = components/operations-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
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

lint-local:
	@echo ${SKIP_STEP_MESSAGE}

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

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(abspath ./bin) go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

copy-crds-to-chart: kustomize
	$(KUSTOMIZE) build config/crd > crds.yaml
	mv crds.yaml ${CHART_PATH}/templates/
