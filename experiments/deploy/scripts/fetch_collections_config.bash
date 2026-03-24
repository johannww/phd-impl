#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
CHAINCODE_RELEASE_NAME="${CHAINCODE_RELEASE_NAME:-${RELEASE_NAME}-chaincode}"
LOCAL_FILE="${LOCAL_FILE:-${SCRIPT_DIR}/../vars/collections_config.json}"

# The ConfigMap name is {{ .Release.Name }}-collections-config
# In deploy.sh, the chaincode release name is defined as ${RELEASE_NAME}-chaincode
CONFIGMAP_NAME="${CHAINCODE_RELEASE_NAME}-collections-config"

echo "Fetching collections_config.json from ConfigMap ${CONFIGMAP_NAME} in namespace ${NAMESPACE}..."

# Ensure the directory exists
mkdir -p "$(dirname "${LOCAL_FILE}")"

# Extract the data from the ConfigMap
kubectl -n "${NAMESPACE}" get configmap "${CONFIGMAP_NAME}" -o jsonpath='{.data.collections_config\.json}' > "${LOCAL_FILE}"

echo "Done. Collection configuration saved to ${LOCAL_FILE}"
