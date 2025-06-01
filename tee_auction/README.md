# Under development

I can re-use modules and data from "github.com/johannww/phd-impl/chaincodes/carbon" and my Microsoft Confidential Container experiments [link](https://github.com/johannww/ubuntu-learning/blob/c79ef19b5f794e165b5ad1df2ca365b92516c3b5/crypto/azure_tee/README.md#L34)
## TODO

I also could use [fabric private chaincodes](https://github.com/hyperledger/fabric-private-chaincode/tree/main/samples) that use Intel SGX

IMPORTANT: sgx has a series of known vulnerabilities (https://sgx.fail/)

Perhaps my idea of using an amd sev is better.

These articles reference it: https://ieeexplore.ieee.org/document/10628912 and 
https://ieeexplore.ieee.org/abstract/document/9049585

# Setup

## Install dependencies and create required Azure resources

```bash
sudo pacman -S azure-cli
az login

# Extension to create the policy for confidential containers
az extension add -n confcom 

# create the Azure Container Registry (ACR) to store the confidential container images
az group create --name carbon --location eastus
REGISTRY_NAME=carbonchain
az acr create --resource-group carbon --name $REGISTRY_NAME --sku Basic --admin-enabled true
az acr login --name $REGISTRY_NAME
```

## Compile the container and push it to the ACR

```bash
REGISTRY=carbonchain.azurecr.io

cd ./tee_auction
docker build -t $REGISTRY/carbon_auction:latest .
docker push $REGISTRY/carbon_auction:latest
```

## Get registry username and password and Generate the policy

```bash
REGISTRY_USER=$(az acr credential show --name $REGISTRY_NAME --query "username")
REGISTRY_PASS=$(az acr credential show --name $REGISTRY_NAME --query "passwords[0].value")

sed -i "s/REGISTRY_USER/$REGISTRY_USER/" ./azure/arm_template.json
sed -i "s/REGISTRY_PASS/$REGISTRY_PASS/" ./azure/arm_template.json

az confcom acipolicygen -a ./azure/arm_template.json

# To the policy in human readable format
az confcom acipolicygen -a ./azure/arm_template.json --outraw-pretty-print
```

## Deploy the confidential container and retrieve the report

```bash
az deployment group create --resource-group carbon --template-file ./azure/arm_template.json

CONTAINER_IP=$(az container show --resource-group carbon --name carbon-auction-container --query "ipAddress.ip" -o tsv)

# Check the report
REPORT=$(curl http://$CONTAINER_IP:8080/report | jq)
echo $REPORT

echo $REPORT | jq '.report_data' | xargs -I{} echo 'The first 32 bytes are the container ed25519 public key: {}'

```

## Teardown container

```bash
az container delete --resource-group carbon --name carbon-auction-container --yes
```

# Verifying the report

To verify the report and assure that the policy generated was satisfied, we can check the `ccePolicy` hash against the one in the ARM template:

```bash
# BOTH SHOULD BE THE SAME
jq '.resources.[0].properties.confidentialComputeProperties.ccePolicy' ./azure/arm_template.json | sed s/\"//g | base64 -d | sha256sum

echo $REPORT | jq '.host_data'
```

We can also re-check that we this policy was the same as generated and posseses the desired container layers:

```bash
az confcom acipolicygen -a ./azure/arm_template.json --outraw-pretty-print
az confcom acipolicygen -a ./azure/arm_template.json
```


