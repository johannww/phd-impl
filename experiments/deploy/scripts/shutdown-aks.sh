#!/usr/bin/env bash
set -euo pipefail

# Shutdown Script for AKS Deployment
# Uninstalls Helm releases and cleans up TEE auction resources
# Does NOT stop the AKS cluster (use experiments/deploy/azure/shutdown_aks.sh for that)

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
CHAINCODE_RELEASE_NAME="${CHAINCODE_RELEASE_NAME:-${RELEASE_NAME}-chaincode}"
TEE_AUCTION_DIR="${TEE_AUCTION_DIR:-${SCRIPT_DIR}/../../../tee_auction}"
RESOURCE_GROUP="${RESOURCE_GROUP:-carbon}"
CLUSTER_NAME="${CLUSTER_NAME:-carbon-aks}"
ENABLE_CLUSTER_MONITORING="${ENABLE_CLUSTER_MONITORING:-true}"

COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${COLOR_BLUE}========================================${NC}"
echo -e "${COLOR_BLUE}AKS Deployment Cleanup${NC}"
echo -e "${COLOR_BLUE}========================================${NC}"
echo ""
echo "Resource Group: ${RESOURCE_GROUP}"
echo "Cluster Name:   ${CLUSTER_NAME}"
echo "Namespace:      ${NAMESPACE}"
echo ""
echo -e "${COLOR_YELLOW}Note: This will NOT stop the AKS cluster.${NC}"
echo -e "${COLOR_YELLOW}To stop the cluster, run: make aks-stop${NC}"
echo ""

# Check if kubectl is configured for the cluster
if ! kubectl cluster-info &>/dev/null; then
    echo -e "${COLOR_RED}Error: kubectl not connected to cluster${NC}"
    echo "Getting AKS credentials..."
    az aks get-credentials --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" --overwrite-existing
fi

# Step 1: Uninstall chaincode Helm release
echo -e "${COLOR_YELLOW}[1/4] Uninstalling chaincode Helm release...${NC}"
if helm status "${CHAINCODE_RELEASE_NAME}" -n "${NAMESPACE}" >/dev/null 2>&1; then
    echo "Uninstalling ${CHAINCODE_RELEASE_NAME}..."
    helm uninstall "${CHAINCODE_RELEASE_NAME}" -n "${NAMESPACE}"
    echo -e "${COLOR_GREEN}✓ Chaincode release uninstalled${NC}\n"
else
    echo "Release ${CHAINCODE_RELEASE_NAME} not found in namespace ${NAMESPACE}, skipping."
    echo ""
fi

# Step 2: Uninstall main Fabric network Helm release
echo -e "${COLOR_YELLOW}[2/4] Uninstalling Fabric network Helm release...${NC}"
if helm status "${RELEASE_NAME}" -n "${NAMESPACE}" >/dev/null 2>&1; then
    echo "Uninstalling ${RELEASE_NAME}..."
    helm uninstall "${RELEASE_NAME}" -n "${NAMESPACE}"
    echo -e "${COLOR_GREEN}✓ Fabric network release uninstalled${NC}\n"
else
    echo "Release ${RELEASE_NAME} not found in namespace ${NAMESPACE}, skipping."
    echo ""
fi

# Step 3: Uninstall monitoring stack
echo -e "${COLOR_YELLOW}[3/4] Uninstalling monitoring stack...${NC}"
if [[ "${ENABLE_CLUSTER_MONITORING}" == "true" ]]; then
    . "${SCRIPT_DIR}/uninstall_monitoring_stack.bash"
else
    echo "Monitoring uninstall disabled (ENABLE_CLUSTER_MONITORING=${ENABLE_CLUSTER_MONITORING})"
fi
echo -e "${COLOR_GREEN}✓ Monitoring cleanup attempted${NC}\n"

# Step 4: Cleanup TEE auction container in Azure
echo -e "${COLOR_YELLOW}[4/4] Cleaning up TEE auction container...${NC}"
if ! make -C "${TEE_AUCTION_DIR}" cleanup 2>/dev/null; then
    echo -e "${COLOR_YELLOW}Warning: TEE auction cleanup failed or container not found (continuing)${NC}"
fi
echo -e "${COLOR_GREEN}✓ TEE auction cleanup attempted${NC}\n"

# Display summary
echo -e "${COLOR_BLUE}========================================${NC}"
echo -e "${COLOR_GREEN}Cleanup Complete!${NC}"
echo -e "${COLOR_BLUE}========================================${NC}"
echo ""
echo "What was cleaned up:"
echo "  - Helm release: ${RELEASE_NAME}"
echo "  - Helm release: ${CHAINCODE_RELEASE_NAME}"
echo "  - TEE auction Azure Container Instance"
echo ""
echo "What remains:"
echo "  - AKS cluster: ${CLUSTER_NAME} (still running)"
echo "  - Kubernetes namespace: ${NAMESPACE}"
echo "  - LoadBalancer public IPs (will be released when cluster is deleted)"
echo ""
echo "Next steps:"
echo ""
echo "  Redeploy:       cd experiments/deploy/scripts && ./deploy-aks.sh --skip-provision"
echo "  Stop cluster:   make aks-stop    (pauses compute billing, keeps data)"
echo "  Delete cluster: make aks-down    (deletes everything including resource group)"
echo ""
