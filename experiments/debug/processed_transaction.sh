#!/bin/bash

# grab transaction id from args
TX_ID="$1"

peer chaincode query \
  -C carbon -n qscc \
  -c "{\"Args\":[\"GetTransactionByID\",\"carbon\",\"${TX_ID}\"]}" 2>/dev/null \
  | head -c -1 \
  | configtxlator proto_decode --type protos.ProcessedTransaction \
  | jq '.transactionEnvelope.payload.data.actions[0].payload.chaincode_proposal_payload.input.chaincode_spec.input.args |= map(@base64d)' --indent 2

