FROM golang:1.23.1 AS builder

COPY . /cc
WORKDIR /cc
RUN go mod tidy && go mod vendor
RUN go build -o carbon ./cmd/carbon/main.go

FROM ubuntu:22.04

COPY --from=builder /cc/carbon /usr/local/bin/carbon
ENTRYPOINT ["/usr/local/bin/carbon"]
EXPOSE 7052

