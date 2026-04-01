#!/bin/bash
set -e

# Default values synced with provision_aks.sh
RESOURCE_GROUP=${RESOURCE_GROUP:-"carbon"}
CLUSTER_NAME=${CLUSTER_NAME:-"carbon-aks"}

echo "Stopping AKS cluster $CLUSTER_NAME in resource group $RESOURCE_GROUP..."
az aks stop --resource-group "$RESOURCE_GROUP" --name "$CLUSTER_NAME"

echo "AKS Cluster $CLUSTER_NAME has been stopped. Compute charges are paused."
