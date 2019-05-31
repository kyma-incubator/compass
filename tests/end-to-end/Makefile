IMG_NAME = compass-end-to-end-test
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)


.PHONY: ci-pr ci-master ci-release resolve build-and-test build-image push-image

ci-pr: resolve build-and-test build-image push-image
ci-master: resolve build-and-test build-image push-image
ci-release: resolve build-and-test build-image push-image

resolve:
	dep ensure -v -vendor-only
build-and-test:
ifndef PROW_JOB_ID
	./before-commit.sh
else
	./before-commit.sh ci
endif

build-image:
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)