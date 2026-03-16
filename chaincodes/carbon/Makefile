.PHONY: cc cc-docker app test test-no-cache docs
REPO := "johannww/phd-exp"

cc:
	go build -o /dev/null ./cmd/carbon/main.go
	
cc-docker:
	go work vendor
	docker build -t ghcr.io/$(REPO)/carbon:latest -f Dockerfile ../..

run:
	CORE_CHAINCODE_ID_NAME=carbon:0.1 CORE_PEER_LOCAL_MSPID=test CHAINCODE_SERVER_ADDRESS=0.0.0.0:7055 go run ./cmd/carbon/main.go

app:
	go build -o /dev/null ./cmd/application/main.go

unit-test:
	go test --tags=testing $(shell go list ./... | grep -v '\.tests/')

test:
	go test --tags=testing ./tests

test-no-amd-network-request:
	go test --tags=testing -skip TestAzureCEEPolicyVerification ./tests

test-no-cache:
	go test --tags=testing -count=1 ./tests

test-network:
	./tests/scripts/setup-test-network.sh
	./tests/scripts/add-idemix-org.sh
	./tests/scripts/test-invoke.sh
	./tests/scripts/cleanup.sh

docs:
	./docs/scripts/generate-diagrams.sh
