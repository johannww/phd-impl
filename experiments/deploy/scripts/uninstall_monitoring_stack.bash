#!/usr/bin/env bash
set -euo pipefail

ENABLE_CLUSTER_MONITORING="${ENABLE_CLUSTER_MONITORING:-true}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
MONITORING_RELEASE_NAME="${MONITORING_RELEASE_NAME:-monitoring}"

if [[ "${ENABLE_CLUSTER_MONITORING}" != "true" ]]; then
  echo "Cluster monitoring disabled (ENABLE_CLUSTER_MONITORING=${ENABLE_CLUSTER_MONITORING}), skipping uninstall."
  return 0 2>/dev/null || exit 0
fi

if helm status "${MONITORING_RELEASE_NAME}" -n "${MONITORING_NAMESPACE}" >/dev/null 2>&1; then
  echo "Uninstalling monitoring stack ${MONITORING_RELEASE_NAME} from ${MONITORING_NAMESPACE}..."
  helm uninstall "${MONITORING_RELEASE_NAME}" -n "${MONITORING_NAMESPACE}"
else
  echo "Monitoring release ${MONITORING_RELEASE_NAME} not found in namespace ${MONITORING_NAMESPACE}, skipping."
fi
