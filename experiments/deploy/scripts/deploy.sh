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
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
CHART_DIR="${CHART_DIR:-${SCRIPT_DIR}/../helm}"
CHAINCODE_CHART_DIR="${CHAINCODE_CHART_DIR:-${CHART_DIR}/charts/chaincode-service}"
CHAINCODE_RELEASE_NAME="${CHAINCODE_RELEASE_NAME:-${RELEASE_NAME}-chaincode}"
TEE_AUCTION_DIR="${TEE_AUCTION_DIR:-${SCRIPT_DIR}/../../../tee_auction}"
CPUS="${CPUS:-6}"
MEMORY="${MEMORY:-12000}"
COLOR_RED='\033[0;31m'
NC='\033[0m' # No Color

(
    echo "Deploying confidential container (tee_auction): docker, policy, deploy..."
    make -C "${TEE_AUCTION_DIR}" docker policy deploy > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo "${COLOR_RED}Error deploying tee_auction confidential container. Please check the output above for details.${NC}"
    fi
) &

. "${SCRIPT_DIR}/install_fabric_binaries.bash"

. "${SCRIPT_DIR}/minikube_setup.sh"

MINIKUBE_IP="$(minikube ip)"

kubectl get namespace "${NAMESPACE}" > /dev/null 2>&1 || kubectl create namespace "${NAMESPACE}"

. "${SCRIPT_DIR}/generate_sicar_cert.bash"

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
  -f <(yq e 'explode(.chaincodeService) | .chaincodeService' ${CHART_DIR}/values.yaml) \
  --set organizationsClaimName="${RELEASE_NAME}-organizations" \
  --set packageConfigMapName="${CHAINCODE_PACKAGE_CONFIGMAP}" \
  "${CC_SET_ARGS[@]}"

. "${SCRIPT_DIR}/fetch_organizations.bash"
. "${SCRIPT_DIR}/fetch_collections_config.bash"

wait

. "${SCRIPT_DIR}/generate_network_profile.bash"

echo "Done."
