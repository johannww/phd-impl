#!/bin/bash
# Environment variables for Org1

SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..
FABRIC_SAMPLES_DIR=fabric-samples
TEST_NETWORK_DIR=fabric-samples/test-network

pushd $TEST_NETWORK_DIR
export PATH=$PATH:$(realpath ../bin)
export FABRIC_CFG_PATH=$(realpath ../config)

pushd addOrg3
./addOrg3.sh up
popd

popd
