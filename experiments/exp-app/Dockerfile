# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy workspace files
COPY go.work go.work.sum ./
COPY chaincodes/go.work.sum chaincodes/go.work.sum

COPY chaincodes/carbon/ ./chaincodes/carbon/
COPY chaincodes/common/ ./chaincodes/common/
COPY chaincodes/interop/ ./chaincodes/interop/
COPY tee_auction/go/ ./tee_auction/go/
COPY data_api/ ./data_api/
COPY experiments/exp-app ./experiments/exp-app
COPY vendor/ ./vendor/
COPY go.work go.work.sum ./

# Copy source code
WORKDIR /build
COPY experiments/exp-app/ ./experiments/exp-app/
COPY chaincodes/common/ ./chaincodes/common/

# Build the binary
WORKDIR /build/experiments/exp-app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /exp-app ./cmd/exp-app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /exp-app-setup ./cmd/exp-app-setup
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /generate-profile ./cmd/generate-profile

# Runtime stage
FROM alpine:3.24.1
LABEL org.opencontainers.image.source https://github.com/johannww/phd-impl

# Install ca-certificates for HTTPS
RUN apk --no-cache --update add ca-certificates bash

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /exp-app .
COPY --from=builder /exp-app-setup .
COPY --from=builder /generate-profile .

# No default command - let Kubernetes control execution
# To run manually: docker run <image> /app/exp-app [args]
# Or: docker run <image> /bin/sh

