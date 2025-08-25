#!/bin/bash
CONTAINER_IP=$(az container show --resource-group carbon --name carbon-auction-container --query "ipAddress.ip" -o tsv)

# Check the report
REPORT=$(curl -sk https://$CONTAINER_IP:8080/report | jq)
echo $REPORT > report.json # save the report
echo "report saved to $(pwd)/report.json"

echo $REPORT | jq '.report_data' | xargs -I{} echo 'The sha512 sum of the TEE self-signed certificate is: {}'

# Check the base64 report
REPORT=$(curl -sk https://$CONTAINER_IP:8080/reportb64)
echo $REPORT > report.txt
echo "base64 report saved to $(pwd)/report.txt"
echo ""

ARM_POLICY_SUM="$(jq '.resources.[0].properties.confidentialComputeProperties.ccePolicy' ./azure/arm_template.json | sed s/\"//g | base64 -d | sha256sum | awk '{print $1}')"
REPORT_POLICY_SUM="$(jq -r '.host_data' report.json | sed s/\"//g)"

if [ "$ARM_POLICY_SUM" == "$REPORT_POLICY_SUM" ]; then
    echo "The ARM policy matches the report policy."
else
    echo "The ARM policy does not match the report policy."
    echo "ARM policy sum: $ARM_POLICY_SUM"
    echo "Report policy sum: $REPORT_POLICY_SUM"
    exit 1
fi

