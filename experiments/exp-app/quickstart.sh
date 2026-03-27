#!/bin/bash
set -euo pipefail

# Quick start script for performance testing

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="${SCRIPT_DIR}/../deploy"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Carbon Credit Blockchain Performance Test ===${NC}"

# Get deployment info
if [ ! -d "${DEPLOY_DIR}/vars/organizations" ]; then
    echo -e "${YELLOW}⚠ Deployment artifacts not found${NC}"
    echo "Please run: cd experiments/deploy/scripts && ./deploy.sh"
    exit 1
fi

# Extract network info
VALUES_FILE="${DEPLOY_DIR}/helm/values.yaml"
MINIKUBE_IP=$(minikube ip)

echo -e "${GREEN}✓ Network Configuration:${NC}"
echo "  Minikube IP: ${MINIKUBE_IP}"
echo "  Channel: carbon"
echo "  Chaincode: carbon"
echo ""

echo -e "${YELLOW}Building binaries...${NC}"
cd "${SCRIPT_DIR}"
make build-all
echo ""

# Generate network profile
echo -e "${GREEN}Generating network profile...${NC}"
PROFILE_FILE="${SCRIPT_DIR}/network-profile.json"
${SCRIPT_DIR}/bin/generate-profile \
    --deploy-dir="${DEPLOY_DIR}" \
    --minikube-ip="${MINIKUBE_IP}" \
    --output="${PROFILE_FILE}" \
    --verbose
echo -e "${GREEN}✓ Network profile saved to: ${PROFILE_FILE}${NC}"
echo ""

# Run performance test
DURATION="10s"
echo -e "${GREEN}Starting performance test (${DURATION} duration)...${NC}"
echo ""

${SCRIPT_DIR}/bin/exp-app \
    --profile="${PROFILE_FILE}" \
    --tps=50 \
    --burst=10 \
    --duration="${DURATION}" \
    --concurrency=5 \
    --metrics-interval=10m \
    --output-json=results.json \
    --output-csv=results.csv

echo ""
echo -e "${GREEN}✓ Test complete!${NC}"
echo "Generated artifacts:"
echo "  - ${PROFILE_FILE} (network profile with all topology info)"
echo "  - results.json (performance metrics summary)"
echo "  - results.csv (detailed transaction log)"
echo ""
echo -e "${GREEN}To review network profile:${NC}"
echo "  jq . ${PROFILE_FILE}"
echo ""
echo -e "${GREEN}To see metrics summary:${NC}"
echo "  jq . results.json"

