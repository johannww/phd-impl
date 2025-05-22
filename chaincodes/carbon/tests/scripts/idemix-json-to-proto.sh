#!/bin/bash
# Environment variables for Org1

ORG_PATH=$1
GO_PROGRAM=$2

# Find all files ending with 'SignerConfig' under ORG_PATH
SIGNER_CONFIGS=$(find "$ORG_PATH" -type f -name '*SignerConfig')

# Convert the list of files to a comma-separated string
SIGNER_CONFIGS_CSV=$(echo "$SIGNER_CONFIGS" | tr '\n' ',')

# Remove trailing comma
SIGNER_CONFIGS_CSV=${SIGNER_CONFIGS_CSV%,}

# Print the comma-separated list
# echo "$SIGNER_CONFIGS_CSV"

go run $GO_PROGRAM --input "$SIGNER_CONFIGS_CSV"

# You can add commands here to process each found file with $GO_PROGRAM if needed
# For example:
# find "$ORG_PATH" -type f -name '*SignerConfig' -exec "$GO_PROGRAM" {} \;

