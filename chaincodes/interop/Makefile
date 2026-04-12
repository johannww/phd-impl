.PHONY: cc cc-docker test test-no-cache docs
REPO := "johannww/phd-impl"

cc:
	go build -o /dev/null ./cmd/interop/main.go
	
cc-docker:
	go work vendor
	docker build -t ghcr.io/$(REPO)/interop:latest -f Dockerfile ../..

unit-test:
	go test $(shell go list ./... | grep -v '\.tests/')

test:
	go test --tags=testing ./tests

test-no-cache:
	go test -count=1 ./tests
