#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/../../vars/kubeconfig"
FABRIC_TAG="${FABRIC_TAG:-3.1.4}"
TOOLS_IMAGE="${TOOLS_IMAGE:-ghcr.io/hyperledger/fabric-tools:${FABRIC_TAG}}"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
CHART_DIR="${CHART_DIR:-${SCRIPT_DIR}/../helm}"
CPUS="${CPUS:-6}"
MEMORY="${MEMORY:-12000}"

MINIKUBE_STATUS="$(minikube status --format '{{.Host}}' 2>/dev/null || true)"
if [[ "${MINIKUBE_STATUS}" == "Running" ]]; then
  echo "Minikube is already running."
elif [[ "${MINIKUBE_STATUS}" == "Stopped" ]]; then
  minikube start
else
  minikube start --embed-certs=true --cpus="${CPUS}" --memory="${MEMORY}"
  mkdir -p "${KUBECONFIG_DIR}"
  cp ~/.kube/config "${KUBECONFIG_DIR}/config"
fi

if minikube image ls | grep -Fq "${TOOLS_IMAGE}"; then
  echo "Image ${TOOLS_IMAGE} already loaded in Minikube, skipping."
else
  echo "Loading ${TOOLS_IMAGE} into Minikube..."
  minikube image load "${TOOLS_IMAGE}"
fi

echo "Installing Helm release ${RELEASE_NAME} in namespace ${NAMESPACE}..."
helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  --create-namespace

echo "Done."
