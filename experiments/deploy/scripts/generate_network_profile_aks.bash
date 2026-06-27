#!/usr/bin/env bash
set -euo pipefail

# Generate network profile for AKS deployment (in-cluster mode)
# Uses Kubernetes DNS for service discovery instead of external IPs

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_ROOT="$(cd -- "$(dirname -- "${SCRIPT_DIR}")" && pwd)"
EXP_APP_DIR="${DEPLOY_ROOT}/../exp-app"
PROFILE_OUTPUT="${PROFILE_OUTPUT:-${DEPLOY_ROOT}/vars/network-profile-aks.json}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"

# Fetch TEE auction container IP from Azure Container Instances (optional)
TEE_IP=""
if command -v az &>/dev/null; then
    TEE_IP="$(az container show \
        --resource-group carbon \
        --name carbon-auction-container \
        --query "ipAddress.ip" \
        -o tsv 2>/dev/null || true)"
fi

echo "Generating AKS network profile (in-cluster mode)..."
echo "Namespace: ${NAMESPACE}"
if [[ -n "${TEE_IP}" ]]; then
    echo "TEE Auction IP: ${TEE_IP}"
fi

# Build generate-profile binary if not exists
if [[ ! -f "${EXP_APP_DIR}/bin/generate-profile" ]]; then
  echo "Building generate-profile binary..."
  cd "${EXP_APP_DIR}"
  make build-all
  cd -
fi

# Generate profile using in-cluster mode (Kubernetes DNS)
TEE_FLAG=""
if [[ -n "${TEE_IP}" ]]; then
    TEE_FLAG="--tee-ip=${TEE_IP}"
fi

"${EXP_APP_DIR}/bin/generate-profile" \
  --deploy-dir="${DEPLOY_ROOT}" \
  --in-cluster \
  --namespace="${NAMESPACE}" \
  ${TEE_FLAG} \
  --output="${PROFILE_OUTPUT}" \
  --verbose

echo ""
echo "AKS network profile generated: ${PROFILE_OUTPUT}"
echo "Profile uses Kubernetes DNS for in-cluster service discovery"

