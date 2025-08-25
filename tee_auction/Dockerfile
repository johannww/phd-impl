# TODO: Finish this file

# This file was based on ./../chaincodes/carbon/Dockerfile
FROM golang:1.24.4 AS builder

COPY ./go /tee_auction/go
WORKDIR /tee_auction/go
RUN if [ ! -d "vendor" ]; then go mod tidy && go mod vendor; fi
RUN go build -o tee_auction ./cmd/auction/main.go

FROM ubuntu:22.04

# This will ensure golang uses the UTC timezone.
# This is important to generate state keys based on the
# timezone.
# See initLocal() in: golang/1.23.1/go/src/time/zoneinfo_unix.go
ENV TZ=UTC

COPY --from=builder /tee_auction/go/tee_auction /usr/local/bin/tee_auction
ENTRYPOINT ["/usr/local/bin/tee_auction"]
EXPOSE 80
EXPOSE 443

