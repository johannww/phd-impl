.PHONY: cc cc-docker test test-no-cache docs
REPO := "johannww/phd-exp"

cc:
	go build -o /dev/null ./cmd/interop/main.go
	
cc-docker:
	go mod vendor && touch vendor/.nosync
	docker build -t ghcr.io/$(REPO)/interop:latest .

unit-test:
	go test $(shell go list ./... | grep -v '\.tests/')

test:
	go test --tags=testing ./tests

test-no-cache:
	go test -count=1 ./tests
