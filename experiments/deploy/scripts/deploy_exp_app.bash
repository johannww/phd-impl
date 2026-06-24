#!/bin/bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-fabric-experiments}"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
EXP_APP_RELEASE_NAME="${EXP_APP_RELEASE_NAME:-${RELEASE_NAME}-exp-app}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_DIR="${CHART_DIR:-${DEPLOY_DIR}/helm}"
EXP_APP_CHART_DIR="${CHART_DIR}/charts/exp-app"

echo "==> Deploying exp-app to namespace: ${NAMESPACE}"

# Generate network profile for in-cluster deployment
echo "==> Generating in-cluster network profile..."
PROFILE_OUTPUT="${DEPLOY_DIR}/vars/network-profile-aks.json"

# Build generate-profile binary if not exists
if [[ ! -f "${DEPLOY_DIR}/../exp-app/bin/generate-profile" ]]; then
  echo "==> Building generate-profile binary..."
  cd "${DEPLOY_DIR}/../exp-app"
  make build-all
  cd -
fi

# Generate profile with --in-cluster flag
"${DEPLOY_DIR}/../exp-app/bin/generate-profile" \
  --deploy-dir="${DEPLOY_DIR}" \
  --in-cluster \
  --namespace="${NAMESPACE}" \
  --output="${PROFILE_OUTPUT}" \
  --verbose

echo "==> Network profile generated at: ${PROFILE_OUTPUT}"

# Create ConfigMap with network profile
echo "==> Creating network-profile ConfigMap..."
kubectl create configmap network-profile \
  --from-file=network-profile.json="${PROFILE_OUTPUT}" \
  -n "${NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -

# Deploy exp-app using Helm with exploded values
echo "==> Deploying exp-app chart..."
helm upgrade --install "${EXP_APP_RELEASE_NAME}" "${EXP_APP_CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  -f <(yq e 'explode(.expApp) | .expApp' ${CHART_DIR}/values.yaml) \
  --set storage.organizationsClaimName="${RELEASE_NAME}-organizations" \
  --set networkProfile.configMapName="network-profile" \
  --wait --timeout 5m

echo ""
echo "==> exp-app deployment complete!"
echo ""
echo "To access the exp-app pod:"
echo "  kubectl exec -it ${EXP_APP_RELEASE_NAME} -n ${NAMESPACE} -- /bin/sh"
echo ""
echo "To run experiments:"
echo "  kubectl exec -it ${EXP_APP_RELEASE_NAME} -n ${NAMESPACE} -- /app/exp-app --profile=/config/network-profile.json --duration=5m --results=/results"
echo ""
echo "To retrieve results:"
echo "  kubectl cp ${NAMESPACE}/${EXP_APP_RELEASE_NAME}:/results ./exp-app-results"
echo ""
