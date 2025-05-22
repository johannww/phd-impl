#!/bin/bash
# This script builds the chaincode and launches it as a service
# in the fabric test network.

DOWNLOAD_PREREQ=$1

FABRIC_VER=2.5.12
FABRIC_CA_VER=1.5.15

# Clone fabric-samples
SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..
FABRIC_SAMPLES_DIR=fabric-samples
TEST_NETWORK_DIR=fabric-samples/test-network
if [ ! -d "$TEST_NETWORK_DIR" ]; then
    git clone git@github.com:hyperledger/fabric-samples.git
fi
touch $FABRIC_SAMPLES_DIR/.nosync

pushd $TEST_NETWORK_DIR
git checkout 5fa5abbbcf
git apply ../../config/idemix-patch.diff

# Install fabric binaries and images
if [[ $DOWNLOAD_PREREQ == "prereq" ]]; then
    ./network.sh prereq -i $FABRIC_VER -cai $FABRIC_CA_VER
fi

# Include fabric binaries in PATH and set FABRIC_CFG_PATH
export PATH=$PATH:$(realpath ../bin)
export FABRIC_CFG_PATH=$(realpath ../config)

# remove existing ccaas. test-network does not do this.
docker container rm -f $(docker container ls -a | grep carbon_ccaas | awk '{print $1}')

./network.sh down -ca
./network.sh up createChannel -ca
./network.sh deployCCAAS -ccn carbon -ccp ../../../
echo 1 > ccversion.txt

popd
