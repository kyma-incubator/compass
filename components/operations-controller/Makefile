APP_NAME = compass-operations-controller
APP_PATH = components/operations-controller
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
VERIFY_IGNORE := /vendor\|/automock\|/api/v1alpha1/zz_generated.deepcopy.go
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
CHART_PATH = $(realpath $(shell pwd)/../..)/chart/compass/charts/operations-controller
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ENVTEST_K8S_VERSION = 1.25.0
ENVTEST ?= $(LOCALBIN)/setup-envtest
KUSTOMIZE_VERSION ?= v3.8.7
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
CONTROLLER_TOOLS_VERSION ?= v0.9.2
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

ifneq ($(strip $(ARTIFACTS)),)
COVERAGE := -coverprofile=$(ARTIFACTS)/filtered.cov; 
endif

test-local: generate manifests envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go $(COVERAGE) test ./... 

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
	$(CONTROLLER_GEN) crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

controller-gen:
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

kustomize: 
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

envtest:
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

copy-crds-to-chart: kustomize
	$(KUSTOMIZE) build config/crd > crds.yaml
	mv crds.yaml ${CHART_PATH}/templates/
