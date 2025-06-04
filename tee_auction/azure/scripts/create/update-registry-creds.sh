#!/bin/bash

REGISTRY_USER=$(az acr credential show --name $REGISTRY_NAME --query "username")
REGISTRY_PASS=$(az acr credential show --name $REGISTRY_NAME --query "passwords[0].value")

sed -i "s/REGISTRY_USER/$REGISTRY_USER/" ./azure/arm_template.json
sed -i "s/REGISTRY_PASS/$REGISTRY_PASS/" ./azure/arm_template.json

az confcom acipolicygen -a ./azure/arm_template.json

# To the policy in human readable format
az confcom acipolicygen -a ./azure/arm_template.json --outraw-pretty-print
