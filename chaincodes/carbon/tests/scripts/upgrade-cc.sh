#!/bin/bash
# This script upgrades the ccaas chaincode

SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..
FABRIC_SAMPLES_DIR=fabric-samples
TEST_NETWORK_DIR=fabric-samples/test-network

pushd $TEST_NETWORK_DIR

# Include fabric binaries in PATH and set FABRIC_CFG_PATH
export PATH=$PATH:$(realpath ../bin)
export FABRIC_CFG_PATH=$(realpath ../config)


# docker container rm -f $(docker container ls -a | grep carbon_ccaas | awk '{print $1}')
CCVERSION=$(cat ccversion.txt)
CCVERSION=$((CCVERSION + 1))
docker container rm -f $(docker container ls -a | grep carbon_ccaas | awk '{print $1}')
./network.sh deployCCAAS -ccn carbon -ccv $CCVERSION -ccp ../../../
echo $CCVERSION > ccversion.txt

popd
