FROM golang:1.24.4 AS builder

COPY . /cc
WORKDIR /cc
RUN if [ ! -d "vendor" ]; then go mod tidy && go mod vendor; fi
RUN go build -o interop ./cmd/interop/main.go

FROM ubuntu:22.04

# This will ensure golang uses the UTC timezone.
# This is important to generate state keys based on the
# timezone.
# See initLocal() in: golang/1.23.1/go/src/time/zoneinfo_unix.go
ENV TZ=UTC

COPY --from=builder /cc/interop /usr/local/bin/interop
ENTRYPOINT ["/usr/local/bin/interop"]
EXPOSE 7052

