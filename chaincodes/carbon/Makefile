.PHONY: cc cc-docker app test test-no-cache docs

cc:
	go build -o /dev/null ./cmd/carbon/main.go
	
cc-docker:
	docker build -t ghcr.io/johannww/phd-impl/carbon:latest .

app:
	go build -o /dev/null ./cmd/application/main.go

unit-test:
	go test --tags=testing $(shell go list ./... | grep -v '\.tests/')

test:
	go test --tags=testing ./tests

test-no-cache:
	go test -count=1 ./tests

test-network:
	./tests/scripts/setup-test-network.sh
	./tests/scripts/add-idemix-org.sh
	./tests/scripts/test-invoke.sh
	./tests/scripts/cleanup.sh

docs:
	./docs/scripts/generate-diagrams.sh
