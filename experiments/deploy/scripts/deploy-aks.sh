#!/usr/bin/env bash
set -euo pipefail

# AKS Deployment Script for BETS Network
# Uses GitHub Container Registry (ghcr.io) for public images
# Usage: ./deploy-aks.sh [--resource-group GROUP]

REPO=${REPO:-"johannww/phd-impl"}
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
FABRIC_TAG="${FABRIC_TAG:-3.1.4}"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
CHART_DIR="${CHART_DIR:-${SCRIPT_DIR}/../helm}"
CHAINCODE_CHART_DIR="${CHAINCODE_CHART_DIR:-${CHART_DIR}/charts/chaincode-service}"
CHAINCODE_RELEASE_NAME="${CHAINCODE_RELEASE_NAME:-${RELEASE_NAME}-chaincode}"
TEE_AUCTION_DIR="${TEE_AUCTION_DIR:-${SCRIPT_DIR}/../../../tee_auction}"

# AKS-specific defaults
RESOURCE_GROUP="${RESOURCE_GROUP:-carbon}"
CLUSTER_NAME="${CLUSTER_NAME:-carbon-aks}"
LOCATION="${LOCATION:-centralindia}"
NODE_COUNT="${NODE_COUNT:-3}"
VM_SIZE="${VM_SIZE:-Standard_D4s_v5}"  # 4 vCPU, 16 GB RAM

COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
NC='\033[0m' # No Color

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --resource-group) RESOURCE_GROUP="$2"; shift 2 ;;
    --cluster-name) CLUSTER_NAME="$2"; shift 2 ;;
    --location) LOCATION="$2"; shift 2 ;;
    --node-count) NODE_COUNT="$2"; shift 2 ;;
    --vm-size) VM_SIZE="$2"; shift 2 ;;
    --skip-provision) SKIP_PROVISION=true; shift ;;
    --skip-push) SKIP_PUSH=true; shift ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--resource-group GROUP] [--cluster-name NAME] [--location LOC] [--skip-provision] [--skip-push]"
      exit 1
      ;;
  esac
done

echo -e "${COLOR_BLUE}========================================${NC}"
echo -e "${COLOR_BLUE}BETS Network Deployment to Azure AKS${NC}"
echo -e "${COLOR_BLUE}========================================${NC}"
echo "Resource Group: ${RESOURCE_GROUP} | Cluster: ${CLUSTER_NAME}"
echo "Location: ${LOCATION} | Nodes: ${NODE_COUNT}x ${VM_SIZE}"
echo "Image Registry: ghcr.io/${REPO}"
echo ""

# Check prerequisites
for cmd in az kubectl helm docker jq yq; do
    if ! command -v $cmd &>/dev/null; then
        echo -e "${COLOR_RED}Error: $cmd not found${NC}"
        exit 1
    fi
done

if ! az account show &>/dev/null; then
    echo -e "${COLOR_RED}Error: Not logged into Azure. Run 'az login'${NC}"
    exit 1
fi

echo -e "${COLOR_GREEN}✓ Prerequisites check passed${NC}\n"

# Step 1: Provision AKS cluster
if [ "${SKIP_PROVISION:-false}" != "true" ]; then
    echo -e "${COLOR_YELLOW}[1/7] Provisioning AKS cluster...${NC}"
    
    # Create resource group
    if ! az group show --name "$RESOURCE_GROUP" &>/dev/null; then
        echo "Creating resource group ${RESOURCE_GROUP}..."
        az group create --name "$RESOURCE_GROUP" --location "$LOCATION"
    fi
    
    # Create or start AKS cluster
    if ! az aks show --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" &>/dev/null; then
        echo "Creating AKS cluster ${CLUSTER_NAME}..."
        az aks create \
            --resource-group "$RESOURCE_GROUP" \
            --name "$CLUSTER_NAME" \
            --location "$LOCATION" \
            --node-count "$NODE_COUNT" \
            --node-vm-size "$VM_SIZE" \
            --enable-managed-identity \
            --generate-ssh-keys \
            --network-plugin azure \
            --load-balancer-sku standard
    else
        CLUSTER_STATE=$(az aks show --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" --query "powerState.code" -o tsv)
        if [ "$CLUSTER_STATE" = "Stopped" ]; then
            echo "Starting AKS cluster..."
            az aks start --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME"
        fi
    fi
    
    echo -e "${COLOR_GREEN}✓ AKS cluster ready${NC}\n"
else
    echo -e "${COLOR_YELLOW}[1/7] Skipping provisioning${NC}\n"
fi

# Get credentials
az aks get-credentials --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" --overwrite-existing

# Step 2: Build and push images to ghcr.io
if [ "${SKIP_PUSH:-false}" != "true" ]; then
    echo -e "${COLOR_YELLOW}[2/7] Building and pushing images to ghcr.io...${NC}"
    echo "Images will be pushed to ghcr.io/${REPO} (public repository)"

    PROJECT_ROOT="${SCRIPT_DIR}/../../.."
    echo "Running 'make docker-push' from project root..."
    make -C "${PROJECT_ROOT}" docker-push

    echo -e "${COLOR_GREEN}✓ Images pushed to ghcr.io/${REPO}${NC}\n"
else
    echo -e "${COLOR_YELLOW}[2/7] Skipping image build and push${NC}\n"
fi

# Step 3: Create namespace (no pull secret needed for public images)
echo -e "${COLOR_YELLOW}[3/7] Creating namespace...${NC}"
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
echo -e "${COLOR_GREEN}✓ Namespace created${NC}\n"

# Step 4: Install Fabric binaries
echo -e "${COLOR_YELLOW}[4/7] Installing Fabric binaries...${NC}"
. "${SCRIPT_DIR}/install_fabric_binaries.bash"
echo -e "${COLOR_GREEN}✓ Fabric binaries installed${NC}\n"

# Step 5: Deploy TEE auction service (in background)
echo -e "${COLOR_YELLOW}[5/7] Deploying TEE auction service...${NC}"
(
    make -C "${TEE_AUCTION_DIR}" resource-group docker policy deploy > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${COLOR_GREEN}✓ TEE auction service deployed${NC}"
    else
        echo -e "${COLOR_RED}Warning: TEE deployment failed${NC}"
    fi
) &
TEE_PID=$!

# Setup SICAR certificates
. "${SCRIPT_DIR}/generate_sicar_cert.bash"
echo -e "${COLOR_GREEN}✓ Certificates ready${NC}\n"

# Step 6: Deploy Fabric network with ghcr.io images
echo -e "${COLOR_YELLOW}[6/7] Deploying Fabric network...${NC}"

helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  --set "network.serviceType=LoadBalancer" \
  --set "storage.organizations.storageClassName=azurefile" \
  --set "images.tools.repository=ghcr.io/${REPO}/fabric-tools" \
  --set "images.tools.tag=${FABRIC_TAG}" \
  --wait --timeout 15m --create-namespace

# Wait for LoadBalancer IPs
echo "Waiting for LoadBalancer external IPs..."
for i in {1..30}; do
    PENDING=$(kubectl get svc -n "${NAMESPACE}" -o json | jq -r '.items[] | select(.spec.type=="LoadBalancer") | select(.status.loadBalancer.ingress == null) | .metadata.name' | wc -l)
    [ "$PENDING" -eq 0 ] && break
    sleep 10
done
echo -e "${COLOR_GREEN}✓ Fabric network deployed${NC}\n"

# Step 7: Deploy chaincodes with ghcr.io images
echo -e "${COLOR_YELLOW}[7/7] Deploying chaincodes...${NC}"
. "${SCRIPT_DIR}/package_chaincodes.bash"

helm upgrade --install "${CHAINCODE_RELEASE_NAME}" "${CHAINCODE_CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  -f <(yq e 'explode(.chaincodeService) | .chaincodeService' ${CHART_DIR}/values.yaml) \
  --set organizationsClaimName="${RELEASE_NAME}-organizations" \
  --set packageConfigMapName="${CHAINCODE_PACKAGE_CONFIGMAP}" \
  --set "chaincodes[0].image.repository=ghcr.io/${REPO}/carbon" \
  --set "chaincodes[1].image.repository=ghcr.io/${REPO}/interop" \
  "${CC_SET_ARGS[@]}" \
  --wait --timeout 10m

echo -e "${COLOR_GREEN}✓ Chaincodes deployed${NC}\n"

# Step 8: Generate network profile
. "${SCRIPT_DIR}/fetch_organizations.bash"
. "${SCRIPT_DIR}/fetch_collections_config.bash"
wait $TEE_PID 2>/dev/null || true
. "${SCRIPT_DIR}/generate_network_profile_aks.bash"
echo -e "${COLOR_GREEN}✓ Network profile generated${NC}\n"

# Display summary
echo -e "${COLOR_BLUE}========================================${NC}"
echo -e "${COLOR_GREEN}Deployment Complete!${NC}"
echo -e "${COLOR_BLUE}========================================${NC}"
echo ""
echo "Next Steps:"
echo ""
echo "  View services:  kubectl get svc -n ${NAMESPACE}"
echo "  View pods:      kubectl get pods -n ${NAMESPACE}"
echo ""
echo "  Run perf test:  cd experiments/exp-app"
echo "                  ./bin/exp-app --profile ../deploy/vars/network-profile.json \\"
echo "                                 --duration 10m --concurrency 20 --enable-metrics"
echo ""
echo "  Pause cluster:  make aks-stop"
echo "  Teardown:       make aks-down RESOURCE_GROUP=${RESOURCE_GROUP}"
echo ""
