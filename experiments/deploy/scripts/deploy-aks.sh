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
LOCATION="${LOCATION:-westcentralus}"
TEE_LOCATION="${TEE_LOCATION:-eastus}" # Azure Confidential Computing is not available in all regions; eastus is a common choice for SEV-SNP support
NODE_COUNT="${NODE_COUNT:-3}"
VM_SIZE="${VM_SIZE:-Standard_D4s_v5}"  # 4 vCPU, 16 GB RAM
ENABLE_IN_CLUSTER_EXP_APP="${ENABLE_IN_CLUSTER_EXP_APP:-true}"
ENABLE_CLUSTER_MONITORING="${ENABLE_CLUSTER_MONITORING:-true}"
MONITORING_SERVICEMONITORS_ENABLED="${MONITORING_SERVICEMONITORS_ENABLED:-${ENABLE_CLUSTER_MONITORING}}"
MONITORING_RELEASE_NAME="${MONITORING_RELEASE_NAME:-monitoring}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
EXP_APP_RESULTS_STORAGE_CLASS="${EXP_APP_RESULTS_STORAGE_CLASS:-azurefile-csi}"

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

if [[ "${ENABLE_CLUSTER_MONITORING}" == "true" ]]; then
    echo -e "${COLOR_YELLOW}[monitoring] Installing kube-prometheus-stack...${NC}"
    . "${SCRIPT_DIR}/install_monitoring_stack.bash"
    echo -e "${COLOR_GREEN}✓ Monitoring stack ready${NC}\n"
fi

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
    make -C "${TEE_AUCTION_DIR}" RESOURCE_GROUP="${RESOURCE_GROUP}" LOCATION="${TEE_LOCATION}" resource-group docker policy deploy > /dev/null 2>&1
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
  --set "monitoring.serviceMonitors.enabled=${MONITORING_SERVICEMONITORS_ENABLED}" \
  --set "monitoring.serviceMonitors.releaseLabel=${MONITORING_RELEASE_NAME}" \
  --set "monitoring.serviceMonitors.namespace=${MONITORING_NAMESPACE}" \
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
  --set "monitoring.serviceMonitors.enabled=${MONITORING_SERVICEMONITORS_ENABLED}" \
  --set "monitoring.serviceMonitors.releaseLabel=${MONITORING_RELEASE_NAME}" \
  --set "monitoring.serviceMonitors.namespace=${MONITORING_NAMESPACE}" \
  "${CC_SET_ARGS[@]}" \
  --wait --timeout 10m

echo -e "${COLOR_GREEN}✓ Chaincodes deployed${NC}\n"

# Fetch organization and collection data
echo "Fetching organization data and collection configs..."
. "${SCRIPT_DIR}/fetch_organizations.bash"
. "${SCRIPT_DIR}/fetch_collections_config.bash"
echo -e "${COLOR_GREEN}✓ Organization data fetched${NC}\n"

# Wait for background TEE deployment process before generating exp-app profile.
# The deployment command can finish before public IP is immediately queryable,
# so deploy_exp_app.bash also retries IP lookup.
if [[ -n "${TEE_PID:-}" ]]; then
  echo -e "${COLOR_YELLOW}Waiting for TEE deployment process to finish...${NC}"
  if wait "${TEE_PID}"; then
    echo -e "${COLOR_GREEN}✓ TEE deployment process finished${NC}\n"
  else
    echo -e "${COLOR_RED}Warning: TEE deployment process failed; exp-app may use mock TEE results${NC}\n"
  fi
fi

# Deploy exp-app pods (in-cluster) or just generate an AKS profile
if [[ "${ENABLE_IN_CLUSTER_EXP_APP}" == "true" ]]; then
    echo -e "${COLOR_YELLOW}Deploying exp-app pods inside cluster (one per peer organization)...${NC}"
    export RELEASE_NAME
    export EXP_APP_RELEASE_NAME="${RELEASE_NAME}-exp-app"
    export EXP_APP_FULLNAME_OVERRIDE="exp-app"
    export PROFILE_OUTPUT="${SCRIPT_DIR}/../vars/network-profile-aks.json"
    export PROFILE_IN_CLUSTER="true"
    export PROFILE_NAMESPACE="${NAMESPACE}"
    export NETWORK_PROFILE_CONFIGMAP_NAME="network-profile"
    export EXP_APP_RESULTS_STORAGE_CLASS
    unset EXP_APP_ORGANIZATION EXP_APP_USER_COUNT EXP_APP_RUN_SETUP EXP_APP_RUN_COUPLED

    . "${SCRIPT_DIR}/deploy_exp_app.bash"
    echo -e "${COLOR_GREEN}✓ exp-app pods deployed${NC}\n"
else
    echo -e "${COLOR_YELLOW}Generating in-cluster AKS network profile only...${NC}"
    . "${SCRIPT_DIR}/generate_network_profile_aks.bash"
    echo -e "${COLOR_GREEN}✓ AKS network profile generated${NC}\n"
fi

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
if [[ "${ENABLE_IN_CLUSTER_EXP_APP}" == "true" ]]; then
    echo "  In-cluster exp-app enabled (one pod per peer organization)."
    echo "  List exp-app pods:"
    echo "    kubectl get pods -n ${NAMESPACE} -l app.kubernetes.io/name=exp-app"
    echo "  Access one pod (example mma):"
    echo "    kubectl exec -it exp-app-mma -n ${NAMESPACE} -- /bin/sh"
    echo "  Run workload in one pod (example mma):"
    echo "    kubectl exec -it exp-app-mma -n ${NAMESPACE} -- /app/exp-app --profile=/config/network-profile.json --duration=5m --output-json=/results/results-mma.json --output-csv=/results/results-mma.csv"
    echo ""
    echo "  Run all experiment pods:"
    echo "    make experiments-run"
    echo ""
    echo "  Retrieve results:"
    echo "    kubectl cp ${NAMESPACE}/exp-app-mma:/results ./exp-app-results-mma"
else
    echo "  In-cluster exp-app disabled. AKS profile generated at experiments/deploy/vars/network-profile-aks.json"
fi
echo ""
echo "  Pause cluster:  make aks-stop"
echo "  Teardown:       make aks-down RESOURCE_GROUP=${RESOURCE_GROUP}"
echo ""
