#!/bin/bash
# Environment variables for Org1

SCRIPT_DIR=$(dirname "$0")
cd $SCRIPT_DIR/..
FABRIC_SAMPLES_DIR=fabric-samples
TEST_NETWORK_DIR=fabric-samples/test-network

pushd $TEST_NETWORK_DIR
export PATH=$PATH:$(realpath ../bin)
export FABRIC_CFG_PATH=$(realpath ../config)

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config
export ORDERER_TLS_ROOTCERT_FILE=${PWD}/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem

# invoke the function
# peer chaincode query -C mychannel -n carbon -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}'
peer chaincode invoke -C mychannel -n carbon -c '{"Args":["CreateSellBid"]}' -o localhost:7050 --tls --cafile $ORDERER_TLS_ROOTCERT_FILE
if [[ $? == 0 ]]; then
  echo "Invoke successful"
else
  echo "Invoke failed"
  exit 1
fi

export CORE_PEER_LOCALMSPID=Org3MSP
export CORE_PEER_LOCALMSPTYPE=idemix
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com
peer chaincode invoke -C mychannel -n carbon -c '{"Args":["CreateSellBid"]}' -o localhost:7050 --tls --cafile $ORDERER_TLS_ROOTCERT_FILE
if [[ $? == 0 ]]; then
  echo "Idemix Invoke successful"
else
  echo "Idemix Invoke failed"
  exit 1
fi
