#!/bin/bash

BLOCK_NUM="$1"

peer channel fetch -c carbon ${BLOCK_NUM} /dev/stdout 2>/dev/null \
| configtxlator proto_decode --type common.Block \
| jq '
  walk(
    if type=="object"
    then with_entries(select(.key!="endorser" and .key!="id_bytes"))
    else .
    end
  )
' \
| jq '
  .data.data[].payload.data.actions[].payload.action.proposal_response_payload.extension.response.payload |= (@base64d | try fromjson catch .)
  |
  .data.data[].payload.data.actions[].payload.chaincode_proposal_payload.input.chaincode_spec.input.args |= map(@base64d)
' --indent 2
