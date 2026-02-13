.PHONY: cc cc-docker test test-no-cache docs

cc:
	go build -o /dev/null ./cmd/interop/main.go
	
cc-docker:
	go mod vendor && touch vendor/.nosync
	docker build -t ghcr.io/johannww/phd-impl/interop:latest .

unit-test:
	go test $(shell go list ./... | grep -v '\.tests/')

test:
	go test --tags=testing ./tests

test-no-cache:
	go test -count=1 ./tests
