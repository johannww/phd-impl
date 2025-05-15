#/usr/bin/env bash

SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..

go run ./cmd/gentestdata/main.go -o data/testdata.json
