APP_NAME = compass-director
APP_PATH = components/director
BUILDPACK = eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.18
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
DEPLOYMENT_DIRECTOR_NAME="compass-director"
DEPLOYMENT_NS_ADAPTER_NAME="compass-ns-adapter"
CRON_JOB_ORD_AGGREGATOR_NAME="compass-ord-aggregator"
CRON_JOB_SYS_FETCHER_NAME="compass-system-fetcher"
CRON_JOB_TEN_FETCHER_ACC_NAME="compass-tenant-fetcher-account-fetcher"
CRON_JOB_TEN_FETCHER_SUB_NAME="compass-tenant-fetcher-subaccount-fetcher"
CRON_JOB_TEN_LOADER_EXT_NAME="compass-director-tenant-loader-external"
JOB_TEN_LOADER_DEF_NAME="compass-director-tenant-loader-default"
JOB_CLNTS_SCOPES_SYNC_NAME="compass-director-clients-scopes-synchronization"
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
	env go build -o bin/tenantfetcher-job ./cmd/tenantfetcher-job/main.go
	env go build -o bin/tenantfetcher-svc ./cmd/tenantfetcher-svc/main.go

install-tools:
	go mod download
	@echo Installing tools from tools.go
	cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

deploy-on-minikube: build-to-minikube
	kubectl config use-context minikube
	kubectl set image -n $(NAMESPACE) deployment/$(DEPLOYMENT_DIRECTOR_NAME) director=compass-director:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/$(DEPLOYMENT_DIRECTOR_NAME)
	kubectl set image -n $(NAMESPACE) deployment/$(DEPLOYMENT_NS_ADAPTER_NAME) ns-adapter=compass-director:latest
	kubectl rollout restart -n $(NAMESPACE) deployment/$(DEPLOYMENT_NS_ADAPTER_NAME)
	kubectl set image -n $(NAMESPACE) cronjob/$(CRON_JOB_ORD_AGGREGATOR_NAME) aggregator=compass-director:latest
	kubectl set image -n $(NAMESPACE) cronjob/$(CRON_JOB_SYS_FETCHER_NAME) system-fetcher=compass-director:latest
	kubectl set image -n $(NAMESPACE) cronjob/$(CRON_JOB_TEN_FETCHER_ACC_NAME) tenant-fetcher=compass-director:latest
	kubectl set image -n $(NAMESPACE) cronjob/$(CRON_JOB_TEN_FETCHER_SUB_NAME) tenant-fetcher=compass-director:latest
	kubectl set image -n $(NAMESPACE) cronjob/$(CRON_JOB_TEN_LOADER_EXT_NAME) loader=compass-director:latest
	kubectl get job -n $(NAMESPACE) $(JOB_TEN_LOADER_DEF_NAME) -o json  | jq "del(.spec.selector)" | jq "del(.spec.template.metadata.labels)" | jq "del(.status)" | kubectl patch -f - --patch '{"spec":  {"template":  {"spec":  {"containers": [{"name": "loader", "image": "compass-director:latest"}]}}}}' --dry-run=client -o yaml | kubectl replace --force -f -
	kubectl get job -n $(NAMESPACE) $(JOB_CLNTS_SCOPES_SYNC_NAME) -o json  | jq "del(.spec.selector)" | jq "del(.spec.template.metadata.labels)" | jq "del(.status)" | kubectl patch -f - --patch '{"spec":  {"template":  {"spec":  {"containers": [{"name": "sync", "image": "compass-director:latest"}]}}}}' --dry-run=client -o yaml | kubectl replace --force -f -
