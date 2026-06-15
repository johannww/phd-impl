REPO := johannww/phd-impl
FABRIC_TAG ?= 3.1.4

.PHONY: docker docker-push

docker:
	FABRIC_TAG=$(FABRIC_TAG) IMAGE_TAG=ghcr.io/$(REPO)/fabric-tools:$(FABRIC_TAG) ./build.sh

docker-push: docker
	docker push ghcr.io/$(REPO)/fabric-tools:$(FABRIC_TAG)
