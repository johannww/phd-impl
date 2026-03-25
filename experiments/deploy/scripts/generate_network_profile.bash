#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_ROOT="$(cd -- "$(dirname -- "${SCRIPT_DIR}")" && pwd)"
EXP_APP_DIR="${DEPLOY_ROOT}/../exp-app"
PROFILE_OUTPUT="${PROFILE_OUTPUT:-${DEPLOY_ROOT}/vars/network-profile.json}"

# Get Minikube IP
MINIKUBE_IP="$(minikube ip)"

echo "Generating network profile..."
go run "${EXP_APP_DIR}/cmd/generate-profile/main.go" \
  --deploy-dir="${DEPLOY_ROOT}" \
  --minikube-ip="${MINIKUBE_IP}" \
  --output="${PROFILE_OUTPUT}" \
  --verbose

echo "Network profile generated: ${PROFILE_OUTPUT}"
