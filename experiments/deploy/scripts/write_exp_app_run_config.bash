#!/usr/bin/env bash
set -euo pipefail

OUTPUT=""
SETUP_POD=""
POD_FLAGS_FILE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      OUTPUT="$2"
      shift 2
      ;;
    --setup-pod)
      SETUP_POD="$2"
      shift 2
      ;;
    --pod-flags-file)
      POD_FLAGS_FILE="$2"
      shift 2
      ;;
    *)
      echo "Usage: $0 --output <file> --setup-pod <pod> --pod-flags-file <file>"
      exit 1
      ;;
  esac
done

if [[ -z "${OUTPUT}" || -z "${SETUP_POD}" || -z "${POD_FLAGS_FILE}" ]]; then
  echo "Usage: $0 --output <file> --setup-pod <pod> --pod-flags-file <file>"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required to write exp_app_flags.json"
  exit 1
fi

per_pod_json='{}'
while IFS=$'\t' read -r pod org coupled; do
  [[ -z "${pod}" ]] && continue
  per_pod_json="$(jq -cn \
    --argjson prev "${per_pod_json}" \
    --arg pod "${pod}" \
    --arg org "${org}" \
    --arg coupled "${coupled}" \
    '$prev + {($pod): {organization: (if $org == "" then null else $org end), run_coupled: ($coupled == "true" or $coupled == "1")}}')"
done < "${POD_FLAGS_FILE}"

jq -n \
  --arg run_id "${RUN_ID:-}" \
  --arg namespace "${NAMESPACE:-}" \
  --arg pod_selector "${POD_SELECTOR:-}" \
  --arg profile_in_pod "${PROFILE_IN_POD:-}" \
  --arg duration "${DURATION:-}" \
  --arg concurrency "${CONCURRENCY:-}" \
  --arg metrics_interval "${METRICS_INTERVAL:-}" \
  --arg mint_interval "${MINT_INTERVAL:-}" \
  --arg buy_bid_interval "${BUY_BID_INTERVAL:-}" \
  --arg sell_bid_interval "${SELL_BID_INTERVAL:-}" \
  --arg auction_interval "${AUCTION_INTERVAL:-}" \
  --arg metrics_formats "${METRICS_FORMATS:-}" \
  --arg tps "${TPS:-}" \
  --arg burst "${BURST:-}" \
  --arg user_count "${USER_COUNT:-}" \
  --arg run_global_setup "${RUN_GLOBAL_SETUP:-}" \
  --arg setup_user_index "${SETUP_USER_INDEX:-}" \
  --arg cluster_metrics_rate_window "${CLUSTER_METRICS_RATE_WINDOW:-}" \
  --arg setup_pod "${SETUP_POD}" \
  --argjson per_pod "${per_pod_json}" \
  '{
    run_id: $run_id,
    kubernetes: {
      namespace: $namespace,
      pod_selector: $pod_selector,
      setup_pod: $setup_pod
    },
    global_setup: {
      run_global_setup: ($run_global_setup == "true" or $run_global_setup == "1"),
      setup_user_index: $setup_user_index
    },
    cluster_metrics: {
      rate_window: (if $cluster_metrics_rate_window == "" then null else $cluster_metrics_rate_window end)
    },
    exp_app_flags: {
      profile: $profile_in_pod,
      duration: $duration,
      concurrency: $concurrency,
      metrics_interval: $metrics_interval,
      mint_interval: $mint_interval,
      buy_bid_interval: $buy_bid_interval,
      sell_bid_interval: $sell_bid_interval,
      auction_interval: $auction_interval,
      enable_metrics: true,
      metrics_formats: $metrics_formats,
      tps: (if $tps == "" then null else $tps end),
      burst: (if $burst == "" then null else $burst end),
      user_count: (if $user_count == "" then null else $user_count end)
    },
    per_pod: $per_pod
  }' > "${OUTPUT}"
