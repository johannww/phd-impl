#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
STOP_MINIKUBE="${STOP_MINIKUBE:-false}"
CHAINCODE_RELEASE_NAME="${CHAINCODE_RELEASE_NAME:-${RELEASE_NAME}-chaincode}"
TEE_AUCTION_DIR="${TEE_AUCTION_DIR:-${SCRIPT_DIR}/../../../tee_auction}"

if helm status "${RELEASE_NAME}" -n "${NAMESPACE}" >/dev/null 2>&1; then
  echo "Uninstalling Helm release ${RELEASE_NAME} from namespace ${NAMESPACE}..."
  helm uninstall "${RELEASE_NAME}" -n "${NAMESPACE}"
else
  echo "Release ${RELEASE_NAME} not found in namespace ${NAMESPACE}, skipping uninstall."
fi

helm uninstall "${CHAINCODE_RELEASE_NAME}" -n "${NAMESPACE}"

echo "Cleaning up tee_auction container resources..."
if ! make -C "${TEE_AUCTION_DIR}" cleanup; then
  echo "tee_auction cleanup failed (continuing shutdown)."
fi

if [[ "${STOP_MINIKUBE}" == "true" ]]; then
  echo "Stopping Minikube..."
  minikube stop
fi

echo "Done."
