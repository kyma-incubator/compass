IMG_NAME = compass-schema-migrator
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

.PHONY: ci-pr ci-master ci-release build-image push-image

ci-pr: validate push-image
ci-master: validate push-image
ci-release: validate push-image

build-image:
	docker build -t $(IMG_NAME) .

validate: build-image
	./validate.sh

push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)