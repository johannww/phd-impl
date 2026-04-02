#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

. ${SCRIPT_DIR}/env.sh

echo "Invoking carbon chaincode (InitLedger) via ${CLUSTER_IP}..."
export FABRIC_CFG_PATH="${SCRIPT_DIR}/../vars"
peer chaincode invoke \
  -o "${ORDERER_ADDRESS}" --cafile "${ORDERER_TLS_CA}" --tls \
  -C "${CHANNEL}" -n carbon \
  ${PEER_ADDRESSES[*]} \
  ${TLS_ROOTCERT_FILES[*]} \
  -c '{"function":"CheckCredAttr","Args":["price_viewer"]}'
