#!/bin/bash
# Extension to create the policy for confidential containers
az extension add -n confcom

# create the Azure Container Registry (ACR) to store the confidential container images
az group create --name carbon --location eastus
az acr create --resource-group carbon --name carbonchain --sku Basic --admin-enabled true
az acr login --name carbonchain

