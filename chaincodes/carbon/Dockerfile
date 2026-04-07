# --- Builder Stage ---
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY chaincodes/go.work.sum chaincodes/go.work.sum

COPY chaincodes/carbon/ ./chaincodes/carbon/
COPY chaincodes/common/ ./chaincodes/common/
COPY chaincodes/interop/ ./chaincodes/interop/
COPY tee_auction/go/ ./tee_auction/go/
COPY data_api/ ./data_api/
COPY experiments/exp-app ./experiments/exp-app
COPY vendor/ ./vendor/

RUN go build -o /usr/local/bin/carbon ./chaincodes/carbon/cmd/carbon

# --- Final Stage ---
FROM ubuntu:24.04

# This will ensure golang uses the UTC timezone.
ENV TZ=UTC

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/carbon /usr/local/bin/carbon
ENTRYPOINT ["/usr/local/bin/carbon"]
EXPOSE 7052
