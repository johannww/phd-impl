.PHONY: build build-all clean run help test fmt docker docker-push

IMAGE_REPO ?= ghcr.io/johannww/phd-impl/exp-app
IMAGE_TAG ?= latest

# Build the performance test binary
build:
	@echo "Building exp-app..."
	go build -o bin/exp-app ./cmd/exp-app
	@echo "Built: bin/exp-app"

# Build all binaries (exp-app and generate-profile)
build-all: build
	@echo "Building generate-profile..."
	go build -o bin/generate-profile ./cmd/generate-profile
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

# Build Docker image
docker:
	@echo "Building Docker image: $(IMAGE_REPO):$(IMAGE_TAG)"
	go work vendor
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) -f Dockerfile ../..
	@echo "Docker image built successfully"

# Build and push Docker image
docker-push: docker
	@echo "Pushing Docker image: $(IMAGE_REPO):$(IMAGE_TAG)"
	docker push $(IMAGE_REPO):$(IMAGE_TAG)
	@echo "Docker image pushed successfully"

# Display help
help:
	@echo "Performance Test Framework Commands:"
	@echo ""
	@echo "  make build              Build the main exp-app binary"
	@echo "  make build-all          Build all binaries (exp-app + generate-profile)"
	@echo "  make clean              Clean build artifacts"
	@echo "  make test               Run tests"
	@echo "  make fmt                Format code"
	@echo "  make docker             Build Docker image"
	@echo "  make docker-push        Build and push Docker image"
	@echo ""
	@echo "Example usage:"
	@echo "  ./bin/generate-profile --deploy-dir=experiments/deploy --output=network.json"
	@echo "  ./bin/exp-app --peer=localhost:7051 --tls-cert=path/to/cert ..."
