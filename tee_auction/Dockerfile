# --- Builder Stage ---
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY chaincodes/go.work.sum chaincodes/go.work.sum

COPY chaincodes/carbon/ ./chaincodes/carbon/
COPY chaincodes/common/ ./chaincodes/common/
COPY chaincodes/interop/ ./chaincodes/interop/
COPY data_api/ ./data_api/
COPY tee_auction/go/ ./tee_auction/go/
COPY experiments/exp-app ./experiments/exp-app
COPY vendor/ ./vendor/

RUN go build -o /usr/local/bin/tee_auction ./tee_auction/go/cmd/auction

# --- Final Stage ---
FROM ubuntu:24.04

# This will ensure golang uses the UTC timezone.
ENV TZ=UTC

COPY --from=builder /usr/local/bin/tee_auction /usr/local/bin/tee_auction
ENTRYPOINT ["/usr/local/bin/tee_auction"]
EXPOSE 80
EXPOSE 443
