#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

NAMESPACE="${NAMESPACE:-fabric-experiments}"
POD_SELECTOR="${POD_SELECTOR:-app.kubernetes.io/name=exp-app}"

RUN_ID="${RUN_ID:-$(date +%Y%m%d-%H%M%S)}"
PROFILE_IN_POD="${PROFILE_IN_POD:-/config/network-profile.json}"
RESULTS_ROOT_IN_POD="${RESULTS_ROOT_IN_POD:-/results}"
LOCAL_RESULTS_BASE="${LOCAL_RESULTS_BASE:-${DEPLOY_DIR}/vars/exp-app-runs}"

DURATION="${DURATION:-5m}"
CONCURRENCY="${CONCURRENCY:-20}"
METRICS_INTERVAL="${METRICS_INTERVAL:-30s}"
METRICS_FORMATS="${METRICS_FORMATS:-png,html,pdf,json}"

TPS="${TPS:-}"
BURST="${BURST:-}"
MINT_INTERVAL="${MINT_INTERVAL:-100ms}"
BUY_BID_INTERVAL="${BUY_BID_INTERVAL:-500ms}"
SELL_BID_INTERVAL="${SELL_BID_INTERVAL:-500ms}"
AUCTION_INTERVAL="${AUCTION_INTERVAL:-15s}"
USER_COUNT="${USER_COUNT:-}"
RUN_GLOBAL_SETUP="${RUN_GLOBAL_SETUP:-true}"
SETUP_USER_INDEX="${SETUP_USER_INDEX:-0}"

echo "==> Running exp-app experiments"
echo "Namespace: ${NAMESPACE}"
echo "Selector: ${POD_SELECTOR}"
echo "Run ID: ${RUN_ID}"

mapfile -t pods < <(kubectl get pods -n "${NAMESPACE}" -l "${POD_SELECTOR}" -o jsonpath='{range .items[?(@.status.phase=="Running")]}{.metadata.name}{"\n"}{end}')
if [[ ${#pods[@]} -eq 0 ]]; then
  echo "Error: no running exp-app pods found with selector ${POD_SELECTOR} in namespace ${NAMESPACE}"
  exit 1
fi

pod_env_value() {
  local pod_name="$1"
  local env_name="$2"
  kubectl get pod "${pod_name}" -n "${NAMESPACE}" -o jsonpath="{range .spec.containers[0].env[?(@.name==\"${env_name}\")]}{.value}{end}"
}

run_pod_experiment() {
  local pod_name="$1"
  local run_coupled_flag="$2"
  local org_name="$3"
  local result_dir="${RESULTS_ROOT_IN_POD}/${RUN_ID}/${pod_name}"

  kubectl exec -n "${NAMESPACE}" "${pod_name}" -- mkdir -p "${result_dir}"

  local cmd=(
    /app/exp-app
    "--profile=${PROFILE_IN_POD}"
    "--duration=${DURATION}"
    "--concurrency=${CONCURRENCY}"
    "--metrics-interval=${METRICS_INTERVAL}"
    "--mint-interval=${MINT_INTERVAL}"
    "--buy-bid-interval=${BUY_BID_INTERVAL}"
    "--sell-bid-interval=${SELL_BID_INTERVAL}"
    "--auction-interval=${AUCTION_INTERVAL}"
    "--output-json=${result_dir}/results.json"
    "--output-csv=${result_dir}/results.csv"
    "--enable-metrics"
    "--metrics-output=${result_dir}/monitoring-exports"
    "--metrics-formats=${METRICS_FORMATS}"
    "--run-coupled=${run_coupled_flag}"
  )

  if [[ -n "${org_name}" ]]; then
    cmd+=("--organization=${org_name}")
  fi
  if [[ -n "${USER_COUNT}" ]]; then
    cmd+=("--user-count=${USER_COUNT}")
  fi
  if [[ -n "${TPS}" ]]; then
    cmd+=("--tps=${TPS}")
  fi
  if [[ -n "${BURST}" ]]; then
    cmd+=("--burst=${BURST}")
  fi

  echo "==> Starting experiment in pod ${pod_name} (org=${org_name:-unknown}, coupled=${run_coupled_flag})"
  kubectl exec -n "${NAMESPACE}" "${pod_name}" -- "${cmd[@]}"
}

setup_pod=""
declare -A pod_org
declare -A pod_run_coupled
for pod in "${pods[@]}"; do
  pod_org["${pod}"]="$(pod_env_value "${pod}" "EXP_APP_ORGANIZATION")"

  run_coupled="$(pod_env_value "${pod}" "EXP_APP_RUN_COUPLED")"
  if [[ -z "${run_coupled}" ]]; then
    run_coupled="false"
  fi
  pod_run_coupled["${pod}"]="${run_coupled}"

  run_setup="$(pod_env_value "${pod}" "EXP_APP_RUN_SETUP")"
  if [[ "${run_setup}" == "true" || "${run_setup}" == "1" ]]; then
    setup_pod="${pod}"
  fi
done

if [[ -z "${setup_pod}" ]]; then
  setup_pod="${pods[0]}"
  echo "==> No setup-designated pod found, defaulting setup owner to ${setup_pod}"
fi

failed_pods=()

if [[ "${RUN_GLOBAL_SETUP}" == "true" ]]; then
  echo "==> Running global setup once via /app/exp-app-setup in pod ${setup_pod}"
  setup_cmd=(
    /app/exp-app-setup
    "--profile=${PROFILE_IN_POD}"
    "--user-index=${SETUP_USER_INDEX}"
  )
  if [[ -n "${pod_org[${setup_pod}]}" ]]; then
    setup_cmd+=("--organization=${pod_org[${setup_pod}]}")
  fi

  if ! kubectl exec -n "${NAMESPACE}" "${setup_pod}" -- "${setup_cmd[@]}"; then
    echo "==> ERROR: global setup failed in pod ${setup_pod}"
    exit 1
  fi
fi

declare -A pid_to_pod
pids=()
for pod in "${pods[@]}"; do
  (
    run_pod_experiment "${pod}" "${pod_run_coupled[${pod}]}" "${pod_org[${pod}]}"
  ) &
  pid="$!"
  pids+=("${pid}")
  pid_to_pod["${pid}"]="${pod}"
done

for pid in "${pids[@]}"; do
  if ! wait "${pid}"; then
    failed_pods+=("${pid_to_pod[${pid}]}")
  fi
done

local_run_dir="${LOCAL_RESULTS_BASE}/${RUN_ID}"
mkdir -p "${local_run_dir}"

download_failed=()
for pod in "${pods[@]}"; do
  remote_path="${RESULTS_ROOT_IN_POD}/${RUN_ID}/${pod}"
  local_path="${local_run_dir}/${pod}"

  rm -rf "${local_path}"
  mkdir -p "${local_path}"

  echo "==> Downloading results from ${pod}:${remote_path}"
  if ! kubectl cp "${NAMESPACE}/${pod}:${remote_path}/." "${local_path}"; then
    download_failed+=("${pod}")
  fi
done

echo
echo "==> Experiment run complete"
echo "Run ID: ${RUN_ID}"
echo "Local results: ${local_run_dir}"

if [[ ${#failed_pods[@]} -gt 0 ]]; then
  echo "Workload failed in pods: ${failed_pods[*]}"
fi

if [[ ${#download_failed[@]} -gt 0 ]]; then
  echo "Download failed for pods: ${download_failed[*]}"
fi

echo "Per-pod artifacts are under: ${local_run_dir}/<pod>/"
echo "Example report: ${local_run_dir}/${setup_pod}/monitoring-exports/report.html"

if [[ ${#failed_pods[@]} -gt 0 || ${#download_failed[@]} -gt 0 ]]; then
  exit 1
fi
