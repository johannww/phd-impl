REPO := johannww/phd-exp
TAG ?= latest

.PHONY: docker test build
build:
	go build ./cmd/server
	go build ./cmd/mockgen

docker:
	docker build -t ghcr.io/$(REPO)/data-api:$(TAG) .

test:
	go test ./... -v
