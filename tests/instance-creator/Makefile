APP_NAME = compass-instance-creator-tests
APP_PATH = tests/instance-creator-tests
BUILDPACK = europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20230323-a1ac96db
SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/scripts
INSTALLATION_SCRIPTS_DIR = $(realpath $(shell pwd)/../..)/installation/scripts
export GO111MODULE = on
export SKIP_STEP_MESSAGE = "Do nothing for Go modules project"
include $(SCRIPTS_DIR)/generic_make_go.mk


test-local:
	@echo ${SKIP_STEP_MESSAGE}

errcheck-local:
	errcheck -blank -asserts -ignoregenerated ./...

e2e-test:
	@$(INSTALLATION_SCRIPTS_DIR)/testing.sh instance-creator

e2e-test-clean:
	kubectl config use-context k3d-kyma
	kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

generate-examples:
	@./generate_examples.sh

sandbox-test:
	@../sandbox.sh instance-creator

run:
	@kubectl exec -n kyma-system oct-tp-compass-e2e-tests-compass-e2e-instance-creator-local-0 -c instance-creator-tests -- ./instance-creator.test -test.run $(testName) -test.v

sandbox-deploy-test:
	env GOOS=linux GOARCH=amd64 go test -c ./tests -o instance-creator.test
	kubectl cp ./instance-creator.test kyma-system/oct-tp-compass-e2e-tests-compass-e2e-instance-creator-local-0:/ -c instance-creator-tests
	rm ./instance-creator.test

sandbox-test-clean:
	kubectl config use-context k3d-kyma
	@kubectl delete testdefinitions.testing.kyma-project.io -n kyma-system compass-e2e-instance-creator-local || true
	@kubectl delete clustertestsuites.testing.kyma-project.io compass-e2e-tests || true

deploy-tests-on-k3d:
	docker build -t ${LOCAL_PREFIX}/compass-e2e-tests:${IMAGE_VERSION} .
	docker push ${LOCAL_PREFIX}/compass-e2e-tests:${IMAGE_VERSION}
	kubectl config use-context k3d-kyma
	$(INSTALLATION_SCRIPTS_DIR)/patch-test-definitions.sh ${IMAGE_VERSION}
