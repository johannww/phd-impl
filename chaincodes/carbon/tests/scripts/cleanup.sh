#!/bin/bash
# This script builds the chaincode and launches it as a service
# in the fabric test network.

FABRIC_VER=2.5.12

SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..
FABRIC_SAMPLES_DIR=fabric-samples
TEST_NETWORK_DIR=fabric-samples/test-network

pushd $TEST_NETWORK_DIR

docker container rm -f $(docker container ls -a | grep carbon_ccaas | awk '{print $1}')
./network.sh down -ca

popd
