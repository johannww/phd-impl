#!/bin/bash

# grab transaction id from args
TX_ID="$1"

peer chaincode query \
  -C carbon -n qscc \
  -c "{\"Args\":[\"GetTransactionByID\",\"carbon\",\"${TX_ID}\"]}" 2>/dev/null \
  | head -c -1 \
  | configtxlator proto_decode --type protos.ProcessedTransaction \
  | jq --indent 2

