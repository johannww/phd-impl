#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
ORG_DIR="${ORG_DIR:-${SCRIPT_DIR}/../../vars/organizations}"

PEER_ORG="org1"
ORDERER_ORG="org2"
CHANNEL="carbon"
PEER_LOCAL_PORT=17051
ORDERER_LOCAL_PORT=17050

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_ID=cli
export CORE_PEER_ADDRESS="localhost:${PEER_LOCAL_PORT}"
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_MSPCONFIGPATH="${ORG_DIR}/peerOrganizations/${PEER_ORG}/users/Admin@${PEER_ORG}/msp"
export CORE_PEER_TLS_ROOTCERT_FILE="${ORG_DIR}/peerOrganizations/${PEER_ORG}/peers/peer0.${PEER_ORG}/tls/ca.crt"
export ORDERER_ADDRESS="localhost:${ORDERER_LOCAL_PORT}"
export ORDERER_TLS_CA="${ORG_DIR}/ordererOrganizations/${ORDERER_ORG}/orderers/orderer0.${ORDERER_ORG}/tls/ca.crt"

cleanup() {
  kill "${PEER_PF_PID}" "${ORDERER_PF_PID}" 2>/dev/null || true
}
trap cleanup EXIT

echo "Port-forwarding peer and orderer..."
kubectl -n "${NAMESPACE}" port-forward "svc/${PEER_ORG}-peer-0" "${PEER_LOCAL_PORT}:7051" &
PEER_PF_PID=$!
kubectl -n "${NAMESPACE}" port-forward "svc/${ORDERER_ORG}-orderer-0" "${ORDERER_LOCAL_PORT}:7050" &
ORDERER_PF_PID=$!

sleep 2

echo "Invoking carbon chaincode (InitLedger)..."
export FABRIC_CFG_PATH="${SCRIPT_DIR}/../../vars"
peer chaincode invoke \
  -o "${ORDERER_ADDRESS}" --cafile "${ORDERER_TLS_CA}" --tls \
  -C "${CHANNEL}" -n carbon \
  --peerAddresses "${CORE_PEER_ADDRESS}" \
  --tlsRootCertFiles "${CORE_PEER_TLS_ROOTCERT_FILE}" \
  -c '{"function":"CheckCredAttr","Args":[]}'
