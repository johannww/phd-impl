#!/usr/bin/env bash
set -euo pipefail

REPO=${REPO:-"johannww/phd-impl"}
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/../vars/kubeconfig"
FABRIC_TAG="${FABRIC_TAG:-3.1.4}"
TOOLS_IMAGE="${TOOLS_IMAGE:-ghcr.io/hyperledger/fabric-tools:${FABRIC_TAG}}"
CARBON_CC_IMAGE="${CARBON_CC_IMAGE:-ghcr.io/$REPO/carbon:latest}"
INTEROP_CC_IMAGE="${INTEROP_CC_IMAGE:-ghcr.io/$REPO/interop:latest}"
SICAR_IMAGE="${SICAR_IMAGE:-ghcr.io/$REPO/data-api:latest}"
CPUS="${CPUS:-6}"
MEMORY="${MEMORY:-12000}"
MINIKUBE_RELOAD_IMAGES="${MINIKUBE_RELOAD_IMAGES:-false}"

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

if ! docker image inspect "${TOOLS_IMAGE}" > /dev/null 2>&1; then
    echo "Building custom ${TOOLS_IMAGE}"
    ${SCRIPT_DIR}/../images/fabric-tools/build.sh
fi

for image in "${TOOLS_IMAGE}" "${CARBON_CC_IMAGE}" "${INTEROP_CC_IMAGE}" "${SICAR_IMAGE}"; do
  if minikube image ls | grep -Fq "${image}" && [ "$MINIKUBE_RELOAD_IMAGES" = "false" ]; then
    echo "Image ${image} already loaded in Minikube, skipping."
  else
    echo "Loading ${image} into Minikube..."
    minikube image load "${image}"
  fi
done
