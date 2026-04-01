#!/bin/bash
set -e

# Default values
RESOURCE_GROUP=${RESOURCE_GROUP:-"carbon"}
LOCATION=${LOCATION:-"centralindia"}
CLUSTER_NAME=${CLUSTER_NAME:-"carbon-aks"}
NODE_COUNT=${NODE_COUNT:-2}
# VM_SIZE=${VM_SIZE:-"Standard_DC2as_v5"} # Standard_DC2as_v5 is SEV-SNP enabled and often cost-effective
VM_SIZE=${VM_SIZE:-"Standard_DS2_v2"}

echo "Checking if resource group $RESOURCE_GROUP exists in $LOCATION..."
if ! az group show --name "$RESOURCE_GROUP" &>/dev/null; then
    echo "Creating resource group $RESOURCE_GROUP..."
    az group create --name "$RESOURCE_GROUP" --location "$LOCATION"
fi

echo "Creating AKS cluster $CLUSTER_NAME with SEV-SNP support..."
# We enable Confidential Computing Addon and use an AMD SEV-SNP enabled VM size
az aks create \
    --resource-group "$RESOURCE_GROUP" \
    --name "$CLUSTER_NAME" \
    --location "$LOCATION" \
    --node-count "$NODE_COUNT" \
    --node-vm-size "$VM_SIZE" \
    --enable-managed-identity \
    --generate-ssh-keys \
    --vm-set-type VirtualMachineScaleSets \
    --network-plugin azure

echo "Fetching credentials for $CLUSTER_NAME..."
az aks get-credentials --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME" --overwrite-existing

echo "AKS Cluster $CLUSTER_NAME is ready."
