#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-fabric-experiments}"
RELEASE_NAME="${RELEASE_NAME:-fabric-experiments}"
EXP_APP_RELEASE_NAME="${EXP_APP_RELEASE_NAME:-${RELEASE_NAME}-exp-app}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_DIR="${CHART_DIR:-${DEPLOY_DIR}/helm}"
EXP_APP_CHART_DIR="${CHART_DIR}/charts/exp-app"
PROFILE_OUTPUT="${PROFILE_OUTPUT:-${DEPLOY_DIR}/vars/network-profile-aks.json}"
PROFILE_IN_CLUSTER="${PROFILE_IN_CLUSTER:-true}"
PROFILE_NAMESPACE="${PROFILE_NAMESPACE:-${NAMESPACE}}"
PROFILE_MINIKUBE_IP="${PROFILE_MINIKUBE_IP:-}"
PROFILE_TEE_IP="${PROFILE_TEE_IP:-}"
TEE_RESOURCE_GROUP="${TEE_RESOURCE_GROUP:-${RESOURCE_GROUP:-carbon}}"
TEE_CONTAINER_NAME="${TEE_CONTAINER_NAME:-carbon-auction-container}"
NETWORK_PROFILE_CONFIGMAP_NAME="${NETWORK_PROFILE_CONFIGMAP_NAME:-network-profile}"
ARM_TEMPLATE_SOURCE="${ARM_TEMPLATE_SOURCE:-${DEPLOY_DIR}/../../tee_auction/azure/arm_template.json}"
EXP_APP_RESULTS_STORAGE_CLASS="${EXP_APP_RESULTS_STORAGE_CLASS:-}"
EXP_APP_FULLNAME_OVERRIDE="${EXP_APP_FULLNAME_OVERRIDE:-}"

echo "==> Deploying exp-app to namespace: ${NAMESPACE}"

if [[ -z "${PROFILE_TEE_IP}" ]] && command -v az &>/dev/null; then
  PROFILE_TEE_IP="$(az container show \
    --resource-group "${TEE_RESOURCE_GROUP}" \
    --name "${TEE_CONTAINER_NAME}" \
    --query "ipAddress.ip" \
    -o tsv 2>/dev/null || true)"
fi

if [[ "${PROFILE_TEE_IP}" == "null" ]]; then
  PROFILE_TEE_IP=""
fi

TEE_FLAG=()
if [[ -n "${PROFILE_TEE_IP}" ]]; then
  echo "==> Using TEE auction IP: ${PROFILE_TEE_IP}"
  TEE_FLAG+=("--tee-ip=${PROFILE_TEE_IP}")
else
  echo "==> TEE lookup target: rg=${TEE_RESOURCE_GROUP}, name=${TEE_CONTAINER_NAME}"
  echo "==> TEE auction IP not found; tee_auction will be disabled in the generated profile"
fi

# Generate network profile
if [[ "${PROFILE_IN_CLUSTER}" == "true" ]]; then
  echo "==> Generating in-cluster network profile..."
else
  echo "==> Generating external network profile..."
fi

# Build generate-profile binary if needed
if [[ ! -f "${DEPLOY_DIR}/../exp-app/bin/generate-profile" ]] || [[ "${DEPLOY_DIR}/../exp-app/cmd/generate-profile/main.go" -nt "${DEPLOY_DIR}/../exp-app/bin/generate-profile" ]] || [[ "${DEPLOY_DIR}/../exp-app/pkg/network/generator.go" -nt "${DEPLOY_DIR}/../exp-app/bin/generate-profile" ]]; then
  echo "==> Building generate-profile binary..."
  cd "${DEPLOY_DIR}/../exp-app"
  make build-all
  cd -
fi

if [[ "${PROFILE_IN_CLUSTER}" == "true" ]]; then
  "${DEPLOY_DIR}/../exp-app/bin/generate-profile" \
    --deploy-dir="${DEPLOY_DIR}" \
    --in-cluster \
    --namespace="${PROFILE_NAMESPACE}" \
    "${TEE_FLAG[@]}" \
    --output="${PROFILE_OUTPUT}" \
    --verbose
else
  if [[ -z "${PROFILE_MINIKUBE_IP}" ]]; then
    echo "Error: PROFILE_MINIKUBE_IP is required when PROFILE_IN_CLUSTER=false"
    exit 1
  fi

  "${DEPLOY_DIR}/../exp-app/bin/generate-profile" \
    --deploy-dir="${DEPLOY_DIR}" \
    --minikube-ip="${PROFILE_MINIKUBE_IP}" \
    "${TEE_FLAG[@]}" \
    --output="${PROFILE_OUTPUT}" \
    --verbose
fi

echo "==> Network profile generated at: ${PROFILE_OUTPUT}"

# Create ConfigMap with network profile
echo "==> Creating network-profile ConfigMap..."
kubectl create configmap "${NETWORK_PROFILE_CONFIGMAP_NAME}" \
  --from-file=network-profile.json="${PROFILE_OUTPUT}" \
  -n "${NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -

# Configure TEE ARM template ConfigMap creation in exp-app chart
TEE_ARM_TEMPLATE_ENABLED="false"
if [[ -f "${ARM_TEMPLATE_SOURCE}" ]]; then
  echo "==> TEE ARM template source found: ${ARM_TEMPLATE_SOURCE}"
  TEE_ARM_TEMPLATE_ENABLED="true"
else
  echo "==> WARNING: ARM template not found at ${ARM_TEMPLATE_SOURCE}; tee ARM template mount disabled"
fi

# Deploy exp-app using Helm with exploded values
echo "==> Deploying exp-app chart..."
EXP_APP_VALUES_FILE="$(mktemp)"
trap 'rm -f "${EXP_APP_VALUES_FILE}"' RETURN
yq e 'explode(.expApp) | .expApp' "${CHART_DIR}/values.yaml" > "${EXP_APP_VALUES_FILE}"

TEE_ARM_TEMPLATE_CONTENT_FLAG=()
if [[ "${TEE_ARM_TEMPLATE_ENABLED}" == "true" ]]; then
  TEE_ARM_TEMPLATE_CONTENT_FLAG=(--set-file "tee.armTemplate.content=${ARM_TEMPLATE_SOURCE}")
fi

RESULTS_STORAGE_CLASS_FLAG=()
if [[ -n "${EXP_APP_RESULTS_STORAGE_CLASS}" ]]; then
  RESULTS_STORAGE_CLASS_FLAG=(--set "storage.results.storageClassName=${EXP_APP_RESULTS_STORAGE_CLASS}")
fi

FULLNAME_OVERRIDE_FLAG=()
if [[ -n "${EXP_APP_FULLNAME_OVERRIDE}" ]]; then
  FULLNAME_OVERRIDE_FLAG=(--set-string "fullnameOverride=${EXP_APP_FULLNAME_OVERRIDE}")
fi

helm upgrade --install "${EXP_APP_RELEASE_NAME}" "${EXP_APP_CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  -f "${EXP_APP_VALUES_FILE}" \
  --set "storage.organizationsClaimName=${RELEASE_NAME}-organizations" \
  --set "networkProfile.configMapName=${NETWORK_PROFILE_CONFIGMAP_NAME}" \
  --set "tee.armTemplate.enabled=${TEE_ARM_TEMPLATE_ENABLED}" \
  "${TEE_ARM_TEMPLATE_CONTENT_FLAG[@]}" \
  "${RESULTS_STORAGE_CLASS_FLAG[@]}" \
  "${FULLNAME_OVERRIDE_FLAG[@]}" \
  --wait --timeout 5m

EXP_APP_POD_NAME="$(kubectl get pods -n "${NAMESPACE}" -l "app.kubernetes.io/instance=${EXP_APP_RELEASE_NAME},app.kubernetes.io/name=exp-app" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
if [[ -z "${EXP_APP_POD_NAME}" ]]; then
  EXP_APP_POD_NAME="${EXP_APP_RELEASE_NAME}"
fi

echo ""
echo "==> exp-app deployment complete!"
echo ""
echo "To access the exp-app pod:"
echo "  kubectl exec -it ${EXP_APP_POD_NAME} -n ${NAMESPACE} -- /bin/sh"
echo ""
echo "To run experiments:"
echo "  kubectl exec -it ${EXP_APP_POD_NAME} -n ${NAMESPACE} -- /app/exp-app --profile=/config/network-profile.json --duration=5m --results=/results"
echo ""
echo "To retrieve results:"
echo "  kubectl cp ${NAMESPACE}/${EXP_APP_POD_NAME}:/results ./exp-app-results"
echo ""
