#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

LOCAL_RESULTS_BASE="${LOCAL_RESULTS_BASE:-${DEPLOY_DIR}/vars/exp-app-runs}"
RUN_ID="${RUN_ID:-}"

if [[ $# -gt 1 ]]; then
  echo "Usage: $0 [run-id|run-dir]"
  exit 1
fi

resolve_run_dir() {
  local arg="${1:-}"

  if [[ -n "${arg}" ]]; then
    if [[ -d "${arg}" ]]; then
      printf '%s\n' "${arg}"
      return 0
    fi
    if [[ -d "${LOCAL_RESULTS_BASE}/${arg}" ]]; then
      printf '%s\n' "${LOCAL_RESULTS_BASE}/${arg}"
      return 0
    fi
    echo "Error: run directory not found for argument: ${arg}" >&2
    return 1
  fi

  if [[ -n "${RUN_ID}" ]]; then
    if [[ -d "${LOCAL_RESULTS_BASE}/${RUN_ID}" ]]; then
      printf '%s\n' "${LOCAL_RESULTS_BASE}/${RUN_ID}"
      return 0
    fi
    echo "Error: run directory not found for RUN_ID=${RUN_ID}" >&2
    return 1
  fi

  local latest
  latest="$(ls -1dt "${LOCAL_RESULTS_BASE}"/* 2>/dev/null | head -n 1 || true)"
  if [[ -z "${latest}" ]]; then
    echo "Error: no run directories found under ${LOCAL_RESULTS_BASE}" >&2
    return 1
  fi

  printf '%s\n' "${latest}"
}

RUN_DIR="$(resolve_run_dir "${1:-}")"
RUN_ID_RESOLVED="$(basename "${RUN_DIR}")"

AGGREGATE_DIR="${RUN_DIR}/aggregate"
mkdir -p "${AGGREGATE_DIR}"

AGGREGATE_CSV="${AGGREGATE_DIR}/results.csv"
AGGREGATE_JSON="${AGGREGATE_DIR}/results.json"
AGGREGATE_MONITORING_DIR="${AGGREGATE_DIR}/monitoring-exports"

python3 "${SCRIPT_DIR}/aggregate_exp_app_results.py" \
  --run-dir "${RUN_DIR}" \
  --output-csv "${AGGREGATE_CSV}" \
  --output-json "${AGGREGATE_JSON}" \
  --run-id "${RUN_ID_RESOLVED}"

monitoring_source=""
for pod_dir in "${RUN_DIR}"/*; do
  if [[ ! -d "${pod_dir}" ]]; then
    continue
  fi
  if [[ "$(basename "${pod_dir}")" == "aggregate" ]]; then
    continue
  fi
  if [[ -d "${pod_dir}/monitoring-exports" ]]; then
    monitoring_source="${pod_dir}/monitoring-exports"
    break
  fi
done

if [[ -n "${monitoring_source}" ]]; then
  rm -rf "${AGGREGATE_MONITORING_DIR}"
  cp -a "${monitoring_source}" "${AGGREGATE_MONITORING_DIR}"
  echo "Monitoring exports copied from: ${monitoring_source}"
else
  echo "Warning: no monitoring-exports directory found in pod results"
fi

echo "==> Aggregated run ${RUN_ID_RESOLVED}"
echo "CSV: ${AGGREGATE_CSV}"
echo "JSON: ${AGGREGATE_JSON}"
if [[ -n "${monitoring_source}" ]]; then
  echo "Monitoring: ${AGGREGATE_MONITORING_DIR}"
fi
