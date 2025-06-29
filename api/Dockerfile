FROM golang:1.24.3 AS builder

COPY . /api
WORKDIR /api
RUN if [ ! -d "vendor" ]; then go mod tidy && go mod vendor; fi
RUN go build -o api ./cmd/api/main.go

FROM ubuntu:22.04

# This will ensure golang uses the UTC timezone.
ENV TZ=UTC

COPY --from=builder /api/api /usr/local/bin/api
ENTRYPOINT ["/usr/local/bin/api"]
EXPOSE 8080

