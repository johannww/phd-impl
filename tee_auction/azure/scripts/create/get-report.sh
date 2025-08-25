#!/bin/bash
CONTAINER_IP=$(az container show --resource-group carbon --name carbon-auction-container --query "ipAddress.ip" -o tsv)

# Check the report
REPORT=$(curl -k https://$CONTAINER_IP:8080/report | jq)
echo $REPORT

echo $REPORT > report.json # save the report

echo $REPORT | jq '.report_data' | xargs -I{} echo 'The sha512 sum of the TEE self-signed certificate is: {}'

# Check the base64 report
REPORT=$(curl -k https://$CONTAINER_IP:8080/reportb64)
echo $REPORT > report.txt
