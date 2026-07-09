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
WALLETS_PER_BUYER="${WALLETS_PER_BUYER:-4}"
RUN_GLOBAL_SETUP="${RUN_GLOBAL_SETUP:-true}"
SETUP_USER_INDEX="${SETUP_USER_INDEX:-0}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
MONITORING_RELEASE_NAME="${MONITORING_RELEASE_NAME:-monitoring}"
CLUSTER_METRICS_STEP="${CLUSTER_METRICS_STEP:-15s}"
CLUSTER_METRICS_RATE_WINDOW="${CLUSTER_METRICS_RATE_WINDOW:-}"

duration_to_seconds() {
  local rest="$1"
  local total_ms=0

  while [[ -n "${rest}" ]]; do
    if [[ "${rest}" =~ ^([0-9]+)(ms|s|m|h)(.*)$ ]]; then
      local value="${BASH_REMATCH[1]}"
      local unit="${BASH_REMATCH[2]}"
      rest="${BASH_REMATCH[3]}"

      case "${unit}" in
        h)
          total_ms=$((total_ms + value * 3600000))
          ;;
        m)
          total_ms=$((total_ms + value * 60000))
          ;;
        s)
          total_ms=$((total_ms + value * 1000))
          ;;
        ms)
          total_ms=$((total_ms + value))
          ;;
      esac
    else
      return 1
    fi
  done

  echo $(((total_ms + 999) / 1000))
}

if [[ -z "${CLUSTER_METRICS_RATE_WINDOW}" ]]; then
  if ! duration_seconds="$(duration_to_seconds "${DURATION}")"; then
    echo "Error: could not parse DURATION=${DURATION}; expected values like 90s, 2m, 1m30s"
    exit 1
  fi

  rate_window_seconds=$((duration_seconds + 60))
  CLUSTER_METRICS_RATE_WINDOW="${rate_window_seconds}s"
fi

echo "==> Running exp-app experiments"
echo "Namespace: ${NAMESPACE}"
echo "Selector: ${POD_SELECTOR}"
echo "Run ID: ${RUN_ID}"
echo "Cluster metrics rate window: ${CLUSTER_METRICS_RATE_WINDOW}"

local_run_dir="${LOCAL_RESULTS_BASE}/${RUN_ID}"
mkdir -p "${local_run_dir}"

POD_FLAGS_FILE="$(mktemp)"
trap 'rm -f "${POD_FLAGS_FILE}"' EXIT

CLUSTER_METRICS_BASELINE="${local_run_dir}/cluster-metrics-baseline.json"
CLUSTER_METRICS_FINAL="${local_run_dir}/cluster-metrics-final.json"
CLUSTER_METRICS_DELTA="${local_run_dir}/cluster-metrics-delta.json"

echo "==> Collecting baseline cluster resource metrics"
"${SCRIPT_DIR}/collect_cluster_resource_metrics.bash" \
  --output "${CLUSTER_METRICS_BASELINE}" \
  --target-namespace "${NAMESPACE}" \
  --monitoring-namespace "${MONITORING_NAMESPACE}" \
  --monitoring-release "${MONITORING_RELEASE_NAME}" \
  --rate-window "${CLUSTER_METRICS_RATE_WINDOW}"

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
  if [[ -n "${WALLETS_PER_BUYER}" ]]; then
    cmd+=("--wallets-per-buyer=${WALLETS_PER_BUYER}")
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
declare -A pod_user_count
for pod in "${pods[@]}"; do
  pod_org["${pod}"]="$(pod_env_value "${pod}" "EXP_APP_ORGANIZATION")"
  pod_user_count["${pod}"]="$(pod_env_value "${pod}" "EXP_APP_USER_COUNT")"

  run_coupled="$(pod_env_value "${pod}" "EXP_APP_RUN_COUPLED")"
  if [[ -z "${run_coupled}" ]]; then
    run_coupled="false"
  fi
  pod_run_coupled["${pod}"]="${run_coupled}"

  run_setup="$(pod_env_value "${pod}" "EXP_APP_RUN_SETUP")"
  if [[ "${run_setup}" == "true" || "${run_setup}" == "1" ]]; then
    setup_pod="${pod}"
  fi

  printf '%s\t%s\t%s\n' "${pod}" "${pod_org[${pod}]}" "${pod_run_coupled[${pod}]}" >> "${POD_FLAGS_FILE}"
done

if [[ -z "${setup_pod}" ]]; then
  setup_pod="${pods[0]}"
  echo "==> No setup-designated pod found, defaulting setup owner to ${setup_pod}"
fi

# set USER_COUNT from pod env if not explicitly set
if [[ -z "${USER_COUNT}" ]]; then
  inferred_user_count="${pod_user_count[${setup_pod}]}"
  if [[ -z "${inferred_user_count}" ]]; then
    for pod in "${pods[@]}"; do
      if [[ -n "${pod_user_count[${pod}]}" ]]; then
        inferred_user_count="${pod_user_count[${pod}]}"
        break
      fi
    done
  fi

  if [[ -n "${inferred_user_count}" ]]; then
    USER_COUNT="${inferred_user_count}"
    echo "==> Using USER_COUNT from pod env (EXP_APP_USER_COUNT): ${USER_COUNT}"
  fi
fi

RUN_FLAGS_JSON="${local_run_dir}/exp_app_flags.json"
RUN_ID="${RUN_ID}" \
NAMESPACE="${NAMESPACE}" \
POD_SELECTOR="${POD_SELECTOR}" \
PROFILE_IN_POD="${PROFILE_IN_POD}" \
DURATION="${DURATION}" \
CONCURRENCY="${CONCURRENCY}" \
METRICS_INTERVAL="${METRICS_INTERVAL}" \
MINT_INTERVAL="${MINT_INTERVAL}" \
BUY_BID_INTERVAL="${BUY_BID_INTERVAL}" \
SELL_BID_INTERVAL="${SELL_BID_INTERVAL}" \
AUCTION_INTERVAL="${AUCTION_INTERVAL}" \
METRICS_FORMATS="${METRICS_FORMATS}" \
TPS="${TPS}" \
BURST="${BURST}" \
USER_COUNT="${USER_COUNT}" \
RUN_GLOBAL_SETUP="${RUN_GLOBAL_SETUP}" \
SETUP_USER_INDEX="${SETUP_USER_INDEX}" \
WALLETS_PER_BUYER="${WALLETS_PER_BUYER}" \
CLUSTER_METRICS_RATE_WINDOW="${CLUSTER_METRICS_RATE_WINDOW}" \
"${SCRIPT_DIR}/write_exp_app_run_config.bash" \
  --output "${RUN_FLAGS_JSON}" \
  --setup-pod "${setup_pod}" \
  --pod-flags-file "${POD_FLAGS_FILE}"

echo "==> Saved run parameters: ${RUN_FLAGS_JSON}"

failed_pods=()

if [[ "${RUN_GLOBAL_SETUP}" == "true" ]]; then
  echo "==> Running global setup once via /app/exp-app-setup in pod ${setup_pod}"
  setup_cmd=(
    /app/exp-app-setup
    "--profile=${PROFILE_IN_POD}"
    "--user-index=${SETUP_USER_INDEX}"
    "--wallets-per-buyer=${WALLETS_PER_BUYER}"
  )
  if [[ -n "${pod_org[${setup_pod}]}" ]]; then
    setup_cmd+=("--organization=${pod_org[${setup_pod}]}")
  fi

  if ! kubectl exec -n "${NAMESPACE}" "${setup_pod}" -- "${setup_cmd[@]}"; then
    echo "==> ERROR: global setup failed in pod ${setup_pod}"
    exit 1
  fi
fi

RUN_START_TS="$(date +%s)"
echo "==> Workload window start (epoch): ${RUN_START_TS}"

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

RUN_END_TS="$(date +%s)"
echo "==> Workload window end (epoch): ${RUN_END_TS}"

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

echo "==> Collecting final cluster resource metrics"
"${SCRIPT_DIR}/collect_cluster_resource_metrics.bash" \
  --output "${CLUSTER_METRICS_FINAL}" \
  --target-namespace "${NAMESPACE}" \
  --monitoring-namespace "${MONITORING_NAMESPACE}" \
  --monitoring-release "${MONITORING_RELEASE_NAME}" \
  --start-ts "${RUN_START_TS}" \
  --end-ts "${RUN_END_TS}" \
  --step "${CLUSTER_METRICS_STEP}" \
  --rate-window "${CLUSTER_METRICS_RATE_WINDOW}"

if command -v jq >/dev/null 2>&1; then
  jq -n \
    --argjson run_start_ts "${RUN_START_TS}" \
    --argjson run_end_ts "${RUN_END_TS}" \
    '
      def deep_subtract($f; $b):
        if ($f | type) == "number" and ($b | type) == "number" then
          ($f - $b)
        elif ($f | type) == "object" and ($b | type) == "object" then
          reduce (((($f | keys_unsorted) + ($b | keys_unsorted)) | unique)[]) as $key ({};
            .[$key] = deep_subtract($f[$key]; $b[$key])
          )
        elif ($f | type) == "object" and ($b | type) == "null" then
          deep_subtract($f; {})
        elif ($f | type) == "null" and ($b | type) == "object" then
          deep_subtract({}; $b)
        elif ($f | type) == "number" and ($b | type) == "null" then
          $f
        elif ($f | type) == "null" and ($b | type) == "number" then
          (0 - $b)
        else
          null
        end;

      (input) as $baseline |
      (input) as $final |
      {
        timestamp: $final.timestamp,
        target_namespace: $final.target_namespace,
        monitoring: $final.monitoring,
        fabric_storage_inventory: ($final.fabric_storage_inventory // null),
        run_window: {
          start_ts: $run_start_ts,
          end_ts: $run_end_ts,
          step: ($final.timeseries.window.step // null)
        },
        baseline: $baseline.metrics,
        final: $final.metrics,
        delta: deep_subtract($final.metrics; $baseline.metrics),
        timeseries: ($final.timeseries // null)
      }
    ' "${CLUSTER_METRICS_BASELINE}" "${CLUSTER_METRICS_FINAL}" > "${CLUSTER_METRICS_DELTA}"
fi

if [[ ${#failed_pods[@]} -gt 0 || ${#download_failed[@]} -gt 0 ]]; then
  exit 1
fi

echo "==> Aggregating cross-pod results for run ${RUN_ID}"
"${SCRIPT_DIR}/aggregate_exp_app_results.bash" "${RUN_ID}"
