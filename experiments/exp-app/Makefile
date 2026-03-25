.PHONY: build build-all clean run help test fmt

# Build the performance test binary
build:
	@echo "Building exp-app..."
	go build -o bin/exp-app ./cmd
	@echo "Built: bin/exp-app"

# Build all binaries (exp-app and generate-profile)
build-all: build
	@echo "Building generate-profile..."
	go build -o bin/generate-profile ./cmd/generate-profile
	go build -o bin/exp-app ./cmd
	@echo "Built: bin/generate-profile"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Display help
help:
	@echo "Performance Test Framework Commands:"
	@echo ""
	@echo "  make build              Build the main exp-app binary"
	@echo "  make build-all          Build all binaries (exp-app + generate-profile)"
	@echo "  make clean              Clean build artifacts"
	@echo "  make test               Run tests"
	@echo "  make fmt                Format code"
	@echo ""
	@echo "Example usage:"
	@echo "  ./bin/generate-profile --deploy-dir=experiments/deploy --output=network.json"
	@echo "  ./bin/exp-app --peer=localhost:7051 --tls-cert=path/to/cert ..."
