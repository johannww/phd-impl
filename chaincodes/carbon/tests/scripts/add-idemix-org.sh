#!/bin/bash
# Environment variables for Org1

SCRIPT_DIR=$(dirname $(realpath "$0"))
cd $SCRIPT_DIR/..
FABRIC_SAMPLES_DIR=fabric-samples
TEST_NETWORK_DIR=fabric-samples/test-network

pushd $TEST_NETWORK_DIR
export PATH=$PATH:$(realpath ../bin)
export FABRIC_CFG_PATH=$(realpath ../config)

pushd addOrg3
ORG_PATH=../organizations/peerOrganizations/org3.example.com
rm -rf $ORG_PATH
./addOrg3.sh generate

echo "Converting the Idemix JSON files to Protobuf format"
$SCRIPT_DIR/idemix-json-to-proto.sh $ORG_PATH $SCRIPT_DIR/../cmd/idemixtoproto/main.go

./addOrg3.sh up
popd

popd
