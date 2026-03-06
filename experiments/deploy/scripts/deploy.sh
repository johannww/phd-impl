#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/../../vars/kubeconfig"
FABRIC_TAG="${FABRIC_TAG:-3.1.4}"
TOOLS_IMAGE="${TOOLS_IMAGE:-ghcr.io/hyperledger/fabric-tools:${FABRIC_TAG}}"
CARBON_CC_IMAGE="${CARBON_CC_IMAGE:-ghcr.io/johannww/phd-impl/carbon:latest}"
INTEROP_CC_IMAGE="${INTEROP_CC_IMAGE:-ghcr.io/johannww/phd-impl/interop:latest}"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
CHART_DIR="${CHART_DIR:-${SCRIPT_DIR}/../helm}"
CHAINCODE_CHART_DIR="${CHAINCODE_CHART_DIR:-${CHART_DIR}/charts/chaincode-service}"
CHAINCODE_RELEASE_NAME="${CHAINCODE_RELEASE_NAME:-${RELEASE_NAME}-chaincode}"
CPUS="${CPUS:-6}"
MEMORY="${MEMORY:-12000}"

. "${SCRIPT_DIR}/install_fabric_binaries.bash"

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

for image in "${TOOLS_IMAGE}" "${CARBON_CC_IMAGE}" "${INTEROP_CC_IMAGE}"; do
  if minikube image ls | grep -Fq "${image}"; then
    echo "Image ${image} already loaded in Minikube, skipping."
  else
    echo "Loading ${image} into Minikube..."
    minikube image load "${image}"
  fi
done

MINIKUBE_IP="$(minikube ip)"

echo "Installing Helm release ${RELEASE_NAME} in namespace ${NAMESPACE}..."
helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  --set "network.externalIPs={${MINIKUBE_IP}}" \
  --wait \
  --create-namespace

. "${SCRIPT_DIR}/package_chaincodes.bash"

echo "Installing Helm release ${CHAINCODE_RELEASE_NAME} in namespace ${NAMESPACE}..."
helm upgrade --install "${CHAINCODE_RELEASE_NAME}" "${CHAINCODE_CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  -f <(yq e '.chaincodeService' ${CHART_DIR}/values.yaml) \
  --set organizationsClaimName="${RELEASE_NAME}-fabric-experiments-organizations" \
  --set packageConfigMapName="${CHAINCODE_PACKAGE_CONFIGMAP}" \
  "${CC_SET_ARGS[@]}"

. "${SCRIPT_DIR}/fetch_organizations.bash"

echo "Done."
