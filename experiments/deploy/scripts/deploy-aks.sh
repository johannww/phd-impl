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
    echo -e "${COLOR_YELLOW}[1/9] Provisioning AKS cluster...${NC}"
    
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
    echo -e "${COLOR_YELLOW}[1/9] Skipping provisioning${NC}\n"
fi

# Get credentials
az aks get-credentials --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" --overwrite-existing

# Step 2: Create namespace first
echo -e "${COLOR_YELLOW}[2/9] Creating namespace...${NC}"
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
echo -e "${COLOR_GREEN}✓ Namespace created${NC}\n"

# Step 3: Create Azure Storage Account for Azure Files (ReadWriteMany PVCs)
echo -e "${COLOR_YELLOW}[3/9] Setting up Azure Files storage...${NC}"

# Ensure Microsoft.Storage provider is registered
STORAGE_PROVIDER_STATE=$(az provider show --namespace Microsoft.Storage --query "registrationState" -o tsv 2>/dev/null || echo "NotRegistered")
if [ "$STORAGE_PROVIDER_STATE" != "Registered" ]; then
    echo "Registering Microsoft.Storage provider..."
    az provider register --namespace Microsoft.Storage
    echo "Waiting for provider registration..."
    while [ "$(az provider show --namespace Microsoft.Storage --query 'registrationState' -o tsv)" != "Registered" ]; do
        sleep 5
    done
    echo "Microsoft.Storage provider registered"
fi

echo -e "${COLOR_GREEN}✓ Azure Files storage configured${NC}\n"

# Step 4: Build and push images to ghcr.io
if [ "${SKIP_PUSH:-false}" != "true" ]; then
    echo -e "${COLOR_YELLOW}[4/9] Building and pushing images to ghcr.io...${NC}"
    echo "Images will be pushed to ghcr.io/${REPO} (public repository)"

    PROJECT_ROOT="${SCRIPT_DIR}/../../.."
    echo "Running 'make docker-rush' from project root..."
    make -C "${PROJECT_ROOT}" docker-push

    echo -e "${COLOR_GREEN}✓ Images pushed to ghcr.io/${REPO}${NC}\n"
else
    echo -e "${COLOR_YELLOW}[4/9] Skipping image build and push${NC}\n"
fi

# Step 5: Install Fabric binaries
echo -e "${COLOR_YELLOW}[5/9] Installing Fabric binaries...${NC}"
. "${SCRIPT_DIR}/install_fabric_binaries.bash"
echo -e "${COLOR_GREEN}✓ Fabric binaries installed${NC}\n"

# Step 6: Deploy TEE auction service (in background)
echo -e "${COLOR_YELLOW}[6/9] Deploying TEE auction service...${NC}"
(
    make -C "${TEE_AUCTION_DIR}" resource-group docker policy deploy > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${COLOR_GREEN}✓ TEE auction service deployed${NC}"
    else
        echo -e "${COLOR_RED}Warning: TEE deployment failed${NC}"
    fi
) &
TEE_PID=$!

# Step 7: Setup SICAR certificates
echo -e "${COLOR_YELLOW}[7/9] Setting up SICAR certificates...${NC}"
. "${SCRIPT_DIR}/generate_sicar_cert.bash"
echo -e "${COLOR_GREEN}✓ Certificates ready${NC}\n"

# Step 8: Deploy Fabric network with ghcr.io images
echo -e "${COLOR_YELLOW}[8/9] Deploying Fabric network...${NC}"

helm upgrade --install "${RELEASE_NAME}" "${CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  --set "network.serviceType=NodePort" \
  --set "storage.organizations.storageClassName=azurefile" \
  --set "images.tools.repository=ghcr.io/${REPO}/fabric-tools" \
  --set "images.tools.tag=${FABRIC_TAG}" \
  --wait --timeout 15m --create-namespace

# # Wait for LoadBalancer IPs
# echo "Waiting for LoadBalancer external IPs..."
# for i in {1..30}; do
#     PENDING=$(kubectl get svc -n "${NAMESPACE}" -o json | jq -r '.items[] | select(.spec.type=="LoadBalancer") | select(.status.loadBalancer.ingress == null) | .metadata.name' | wc -l)
#     [ "$PENDING" -eq 0 ] && break
#     sleep 10
# done
echo -e "${COLOR_GREEN}✓ Fabric network deployed${NC}\n"

# Step 9: Deploy chaincodes with ghcr.io images
echo -e "${COLOR_YELLOW}[9/9] Deploying chaincodes...${NC}"
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

# Fetch organization and collection data
echo "Fetching organization data and collection configs..."
. "${SCRIPT_DIR}/fetch_organizations.bash"
. "${SCRIPT_DIR}/fetch_collections_config.bash"
echo -e "${COLOR_GREEN}✓ Organization data fetched${NC}\n"

# Deploy exp-app pod (in-cluster)
echo -e "${COLOR_YELLOW}Deploying exp-app pod...${NC}"
export RELEASE_NAME EXP_APP_RELEASE_NAME="${RELEASE_NAME}-exp-app"
. "${SCRIPT_DIR}/deploy_exp_app.bash"
echo -e "${COLOR_GREEN}✓ exp-app pod deployed${NC}\n"

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
echo "  Access exp-app: kubectl exec -it ${RELEASE_NAME}-exp-app -n ${NAMESPACE} -- /bin/sh"
echo ""
echo "  Run experiments:"
echo "    kubectl exec -it ${RELEASE_NAME}-exp-app -n ${NAMESPACE} -- /app/exp-app \\"
echo "      --profile=/config/network-profile.json \\"
echo "      --duration=5m \\"
echo "      --concurrency=20 \\"
echo "      --enable-metrics \\"
echo "      --results=/results"
echo ""
echo "  Retrieve results:"
echo "    kubectl cp ${NAMESPACE}/${RELEASE_NAME}-exp-app:/results ./exp-app-results"
echo ""
echo "  Pause cluster:  make aks-stop"
echo "  Teardown:       make aks-down RESOURCE_GROUP=${RESOURCE_GROUP}"
echo ""
