IMG_NAME = compass-schema-migrator
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

.PHONY: ci-pr ci-master ci-release build-image push-image

ci-pr: build-image validate push-image
ci-master: build-image validate push-image
ci-release: build-image validate push-image

build-image:
	docker build -t $(IMG_NAME) .

validate:
	./validate.sh

push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)