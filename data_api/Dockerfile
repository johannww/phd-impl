FROM golang:1.26 AS builder

WORKDIR /data_api
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server
RUN CGO_ENABLED=0 go build -o mockgen ./cmd/mockgen

FROM ubuntu:22.04

ENV TZ=UTC

COPY --from=builder /data_api/server  /usr/local/bin/server
COPY --from=builder /data_api/mockgen /usr/local/bin/mockgen
ENTRYPOINT ["/usr/local/bin/server"]
EXPOSE 8443
