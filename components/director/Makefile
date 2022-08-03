APP_NAME = compass-director
APP_PATH = components/director
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
NAMESPACE="compass-system"
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"

include $(SCRIPTS_DIR)/generic_make_go.mk
VERIFY_IGNORE := /vendor\|/automock\|/testdata

verify:: gqlgen check-gqlgen

errcheck-local:
	errcheck -blank -asserts -exclude ./excluded.txt -ignoregenerated ./...

resolve-local:
	@echo ${SKIP_STEP_MESSAGE}

ensure-local:
	@echo ${SKIP_STEP_MESSAGE}

dep-status-local:
	@echo ${SKIP_STEP_MESSAGE}

build-local:
	env go build -o bin/director ./cmd/director/main.go
	env go build -o bin/tenantfetcher-svc ./cmd/tenantfetcher-svc/main.go

install-tools:
	go mod download
	@echo Installing tools from tools.go
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

deploy-on-k3d: build-for-k3d
	kubectl config use-context k3d-kyma
	kubectl patch -n $(NAMESPACE) deployment/compass-director -p '{"spec":{"template":{"spec":{"containers":[{"name":"director","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-director director=k3d-kyma-registry:5001/compass-director:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-director
	kubectl patch -n $(NAMESPACE) deployment/compass-ns-adapter -p '{"spec":{"template":{"spec":{"containers":[{"name":"ns-adapter","imagePullPolicy":"Always"}]}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-ns-adapter ns-adapter=k3d-kyma-registry:5001/compass-director:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-ns-adapter
	kubectl patch -n $(NAMESPACE) cronjob/compass-ord-aggregator -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"aggregator","imagePullPolicy":"Always"}]}}}}}}'
	kubectl set image -n $(NAMESPACE) deployment/compass-tenant-fetcher tenant-fetcher=k3d-kyma-registry:5001/compass-director:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-tenant-fetcher
	kubectl set image -n $(NAMESPACE) deployment/compass-tenant-fetcher tenant-fetcher=k3d-kyma-registry:5001/compass-director:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/compass-tenant-fetcher
	kubectl set image -n $(NAMESPACE) cronjob/compass-ord-aggregator aggregator=k3d-kyma-registry:5001/compass-director:latest
	kubectl patch -n $(NAMESPACE) cronjob/compass-system-fetcher -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"system-fetcher","imagePullPolicy":"Always"}]}}}}}}'
	kubectl set image -n $(NAMESPACE) cronjob/compass-system-fetcher system-fetcher=k3d-kyma-registry:5001/compass-director:latest
	kubectl patch -n $(NAMESPACE) cronjob/compass-director-tenant-loader-external -p '{"spec":{"jobTemplate":{"spec":{"template":{"spec":{"containers":[{"name":"loader","imagePullPolicy":"Always"}]}}}}}}'
	kubectl set image -n $(NAMESPACE) cronjob/compass-director-tenant-loader-external loader=k3d-kyma-registry:5001/compass-director:latest
	kubectl get job -n $(NAMESPACE) compass-director-tenant-loader-default -o json  | jq "del(.spec.selector)" | jq "del(.spec.template.metadata.labels)" | jq "del(.status)" | kubectl patch -f - --patch '{"spec":  {"template":  {"spec":  {"containers": [{"name": "loader", "image": "k3d-kyma-registry:5001/compass-director:latest"}]}}}}' --dry-run=client -o yaml | kubectl replace --force -f -
	kubectl get job -n $(NAMESPACE) compass-director-clients-scopes-synchronization -o json  | jq "del(.spec.selector)" | jq "del(.spec.template.metadata.labels)" | jq "del(.status)" | kubectl patch -f - --patch '{"spec":  {"template":  {"spec":  {"containers": [{"name": "sync", "image": "k3d-kyma-registry:5001/compass-director:latest"}]}}}}' --dry-run=client -o yaml | kubectl replace --force -f -
