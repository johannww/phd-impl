FROM golang:1.24.3 AS builder

COPY . /cc
WORKDIR /cc
RUN if [ ! -d "vendor" ]; then go mod tidy && go mod vendor; fi
RUN go build -o carbon ./cmd/carbon/main.go

FROM ubuntu:22.04

# This will ensure golang uses the UTC timezone.
# This is important to generate state keys based on the
# timezone.
# See initLocal() in: golang/1.23.1/go/src/time/zoneinfo_unix.go
ENV TZ=UTC

COPY --from=builder /cc/carbon /usr/local/bin/carbon
ENTRYPOINT ["/usr/local/bin/carbon"]
EXPOSE 7052

