#!/bin/bash

BLOCK_NUM="$1"

peer channel fetch -c carbon ${BLOCK_NUM} /dev/stdout 2>/dev/null \
  | configtxlator proto_decode --type common.Block \
  | jq --indent 2

