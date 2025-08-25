#!/bin/bash
CONTAINER_IP=$(az container show --resource-group carbon --name carbon-auction-container --query "ipAddress.ip" -o tsv)

# Check the report
REPORT=$(curl http://$CONTAINER_IP:8080/report | jq)
echo $REPORT

echo $REPORT > report.json # save the report

echo $REPORT | jq '.report_data' | xargs -I{} echo 'The first 32 bytes are the container ed25519 public key: {}'

# Check the base64 report
REPORT=$(curl http://$CONTAINER_IP:8080/reportb64)
echo $REPORT > report.txt
