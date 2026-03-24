#/usr/bin/env bash

SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..

go run ./cmd/genserializedauctiondata/main.go -o data/testdata_tee.json --format json
go run ./cmd/genserializedauctiondata/main.go -o data/testdata_tee.pb --format proto
