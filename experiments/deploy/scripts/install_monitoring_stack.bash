#!/usr/bin/env bash
set -euo pipefail

ENABLE_CLUSTER_MONITORING="${ENABLE_CLUSTER_MONITORING:-true}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
MONITORING_RELEASE_NAME="${MONITORING_RELEASE_NAME:-monitoring}"
MONITORING_GRAFANA_ENABLED="${MONITORING_GRAFANA_ENABLED:-false}"
MONITORING_ALERTMANAGER_ENABLED="${MONITORING_ALERTMANAGER_ENABLED:-false}"
MONITORING_TIMEOUT="${MONITORING_TIMEOUT:-10m}"

if [[ "${ENABLE_CLUSTER_MONITORING}" != "true" ]]; then
  echo "Cluster monitoring disabled (ENABLE_CLUSTER_MONITORING=${ENABLE_CLUSTER_MONITORING}), skipping installation."
  return 0 2>/dev/null || exit 0
fi

echo "Installing kube-prometheus-stack (${MONITORING_RELEASE_NAME}) in namespace ${MONITORING_NAMESPACE}..."

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --force-update >/dev/null
helm repo update prometheus-community >/dev/null

kubectl create namespace "${MONITORING_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f - >/dev/null

helm upgrade --install "${MONITORING_RELEASE_NAME}" prometheus-community/kube-prometheus-stack \
  --namespace "${MONITORING_NAMESPACE}" \
  --set "grafana.enabled=${MONITORING_GRAFANA_ENABLED}" \
  --set "alertmanager.enabled=${MONITORING_ALERTMANAGER_ENABLED}" \
  --wait --timeout "${MONITORING_TIMEOUT}"

echo "kube-prometheus-stack installed: release=${MONITORING_RELEASE_NAME}, namespace=${MONITORING_NAMESPACE}"
