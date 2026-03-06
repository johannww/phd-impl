#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ORG_DIR="${ORG_DIR:-${SCRIPT_DIR}/../vars/organizations}"

PEER_ORG="org1"
ORDERER_ORG="org2"
CHANNEL="carbon"
# NodePort values must match values.yaml nodePortBase
PEER_NODE_PORT=30060
ORDERER_NODE_PORT=30050

CLUSTER_IP="$(minikube ip)"

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_ID=cli
export CORE_PEER_ADDRESS="${CLUSTER_IP}:${PEER_NODE_PORT}"
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_MSPCONFIGPATH="${ORG_DIR}/peerOrganizations/${PEER_ORG}/users/Admin@${PEER_ORG}/msp"
export CORE_PEER_TLS_ROOTCERT_FILE="${ORG_DIR}/peerOrganizations/${PEER_ORG}/peers/peer0.${PEER_ORG}/tls/ca.crt"
export ORDERER_ADDRESS="${CLUSTER_IP}:${ORDERER_NODE_PORT}"
export ORDERER_TLS_CA="${ORG_DIR}/ordererOrganizations/${ORDERER_ORG}/orderers/orderer0.${ORDERER_ORG}/tls/ca.crt"

echo "Invoking carbon chaincode (InitLedger) via ${CLUSTER_IP}..."
export FABRIC_CFG_PATH="${SCRIPT_DIR}/../vars"
peer chaincode invoke \
  -o "${ORDERER_ADDRESS}" --cafile "${ORDERER_TLS_CA}" --tls \
  -C "${CHANNEL}" -n carbon \
  --peerAddresses "${CORE_PEER_ADDRESS}" \
  --tlsRootCertFiles "${CORE_PEER_TLS_ROOTCERT_FILE}" \
  -c '{"function":"CheckCredAttr","Args":["admin"]}'
