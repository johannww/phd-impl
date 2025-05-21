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
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/User2@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config
export ORDERER_TLS_ROOTCERT_FILE=${PWD}/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem

popd

cd ..

APP_CMD_MAIN=./cmd/application/main.go
tree $CORE_PEER_MSPCONFIGPATH
# TODO: this is not working yet
go run $APP_CMD_MAIN --idemix true --mspId Org2MSP --mspPath $CORE_PEER_MSPCONFIGPATH
