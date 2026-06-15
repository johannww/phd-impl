REPO := johannww/phd-impl
TAG ?= latest

.PHONY: docker docker-push test build
build:
	go build ./cmd/server
	go build ./cmd/mockgen

docker:
	docker build -t ghcr.io/$(REPO)/data-api:$(TAG) .

docker-push: docker
	docker push ghcr.io/$(REPO)/data-api:$(TAG)

test:
	go test ./...
