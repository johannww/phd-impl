#!/usr/bin/env bash
set -euo pipefail

OUTPUT=""
TARGET_NAMESPACE="${TARGET_NAMESPACE:-fabric-experiments}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
MONITORING_RELEASE_NAME="${MONITORING_RELEASE_NAME:-monitoring}"
START_TS=""
END_TS=""
STEP="15s"
RATE_WINDOW="${RATE_WINDOW:-5m}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      OUTPUT="$2"
      shift 2
      ;;
    --target-namespace)
      TARGET_NAMESPACE="$2"
      shift 2
      ;;
    --monitoring-namespace)
      MONITORING_NAMESPACE="$2"
      shift 2
      ;;
    --monitoring-release)
      MONITORING_RELEASE_NAME="$2"
      shift 2
      ;;
    --start-ts)
      START_TS="$2"
      shift 2
      ;;
    --end-ts)
      END_TS="$2"
      shift 2
      ;;
    --step)
      STEP="$2"
      shift 2
      ;;
    --rate-window)
      RATE_WINDOW="$2"
      shift 2
      ;;
    *)
      echo "Usage: $0 --output <file> [--target-namespace ns] [--monitoring-namespace ns] [--monitoring-release name] [--start-ts epoch] [--end-ts epoch] [--step duration] [--rate-window duration]"
      exit 1
      ;;
  esac
done

if [[ -z "${OUTPUT}" ]]; then
  echo "Error: --output is required"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required"
  exit 1
fi

if [[ -n "${START_TS}" && -z "${END_TS}" ]]; then
  echo "Error: --end-ts is required when --start-ts is set"
  exit 1
fi

if [[ -z "${START_TS}" && -n "${END_TS}" ]]; then
  echo "Error: --start-ts is required when --end-ts is set"
  exit 1
fi

PROM_SERVICE="${MONITORING_RELEASE_NAME}-kube-prometheus-prometheus"
PROM_QUERY_BASE="/api/v1/namespaces/${MONITORING_NAMESPACE}/services/http:${PROM_SERVICE}:9090/proxy/api/v1/query"
PROM_QUERY_RANGE_BASE="/api/v1/namespaces/${MONITORING_NAMESPACE}/services/http:${PROM_SERVICE}:9090/proxy/api/v1/query_range"

encode_uri() {
  jq -nr --arg value "$1" '$value|@uri'
}

query_prometheus_raw() {
  local query="$1"
  local encoded
  encoded="$(encode_uri "${query}")"
  kubectl get --raw "${PROM_QUERY_BASE}?query=${encoded}" 2>/dev/null || true
}

query_prometheus_raw_range() {
  local query="$1"
  local encoded_query encoded_start encoded_end encoded_step
  encoded_query="$(encode_uri "${query}")"
  encoded_start="$(encode_uri "${START_TS}")"
  encoded_end="$(encode_uri "${END_TS}")"
  encoded_step="$(encode_uri "${STEP}")"
  kubectl get --raw "${PROM_QUERY_RANGE_BASE}?query=${encoded_query}&start=${encoded_start}&end=${encoded_end}&step=${encoded_step}" 2>/dev/null || true
}

query_prometheus() {
  local query="$1"
  local response

  response="$(query_prometheus_raw "${query}")"
  if [[ -z "${response}" ]]; then
    echo "null"
    return 0
  fi

  echo "${response}" | jq -r 'if .status == "success" and (.data.result | length) > 0 then (.data.result[0].value[1] | tonumber?) // null else null end'
}

query_prometheus_map() {
  local query="$1"
  local key_label="$2"
  local response

  response="$(query_prometheus_raw "${query}")"
  if [[ -z "${response}" ]]; then
    echo '{}'
    return 0
  fi

  echo "${response}" | jq -c --arg key_label "${key_label}" '
    if .status == "success" then
      reduce .data.result[] as $row ({};
        .[($row.metric[$key_label] // "unknown")] = (($row.value[1] | tonumber?) // null)
      )
    else
      {}
    end'
}

query_prometheus_range_values() {
  local query="$1"
  local response

  if [[ -z "${START_TS}" ]]; then
    echo '[]'
    return 0
  fi

  response="$(query_prometheus_raw_range "${query}")"
  if [[ -z "${response}" ]]; then
    echo '[]'
    return 0
  fi

  echo "${response}" | jq -c '
    if .status == "success" and (.data.result | length) > 0 then
      (.data.result[0].values // [] | map({ts: (.[0] | tonumber), value: ((.[1] | tonumber?) // null)}))
    else
      []
    end'
}

query_prometheus_range_grouped() {
  local query="$1"
  local key_label="$2"
  local response

  if [[ -z "${START_TS}" ]]; then
    echo '{}'
    return 0
  fi

  response="$(query_prometheus_raw_range "${query}")"
  if [[ -z "${response}" ]]; then
    echo '{}'
    return 0
  fi

  echo "${response}" | jq -c --arg key_label "${key_label}" '
    if .status == "success" then
      reduce .data.result[] as $row ({};
        .[($row.metric[$key_label] // "unknown")] = (($row.values // []) | map({ts: (.[0] | tonumber), value: ((.[1] | tonumber?) // null)}))
      )
    else
      {}
    end'
}

peer_pod_regex=".*-peer-.*"
orderer_pod_regex=".*-orderer-.*"
exp_app_pod_regex="^exp-app-.*"
chaincode_pod_regex="^(carbon|interop)-.*"

cpu_namespace_query="rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod!=\"\"}[${RATE_WINDOW}])"
memory_namespace_query="container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod!=\"\"}"

peer_cpu_total_query="sum(rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${peer_pod_regex}\"}[${RATE_WINDOW}]))"
orderer_cpu_total_query="sum(rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${orderer_pod_regex}\"}[${RATE_WINDOW}]))"
exp_app_cpu_total_query="sum(rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${exp_app_pod_regex}\"}[${RATE_WINDOW}]))"
chaincode_cpu_total_query="sum(rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${chaincode_pod_regex}\"}[${RATE_WINDOW}]))"

peer_memory_total_query="sum(container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${peer_pod_regex}\"})"
orderer_memory_total_query="sum(container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${orderer_pod_regex}\"})"
exp_app_memory_total_query="sum(container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${exp_app_pod_regex}\"})"
chaincode_memory_total_query="sum(container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${chaincode_pod_regex}\"})"

peer_cpu_per_pod_query="sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${peer_pod_regex}\"}[${RATE_WINDOW}]))"
orderer_cpu_per_pod_query="sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${orderer_pod_regex}\"}[${RATE_WINDOW}]))"
exp_app_cpu_per_pod_query="sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${exp_app_pod_regex}\"}[${RATE_WINDOW}]))"
chaincode_cpu_per_pod_query="sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${chaincode_pod_regex}\"}[${RATE_WINDOW}]))"

peer_memory_per_pod_query="sum by (pod) (container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${peer_pod_regex}\"})"
orderer_memory_per_pod_query="sum by (pod) (container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${orderer_pod_regex}\"})"
exp_app_memory_per_pod_query="sum by (pod) (container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${exp_app_pod_regex}\"})"
chaincode_memory_per_pod_query="sum by (pod) (container_memory_working_set_bytes{namespace=\"${TARGET_NAMESPACE}\",pod=~\"${chaincode_pod_regex}\"})"

cluster_cpu_cores_used="$(query_prometheus "sum(rate(container_cpu_usage_seconds_total{pod!=\"\"}[${RATE_WINDOW}]))")"
cluster_memory_working_set_bytes="$(query_prometheus 'sum(container_memory_working_set_bytes{pod!=""})')"
cluster_network_receive_bytes_per_sec="$(query_prometheus "sum(rate(container_network_receive_bytes_total{pod!=\"\"}[${RATE_WINDOW}]))")"
cluster_network_transmit_bytes_per_sec="$(query_prometheus "sum(rate(container_network_transmit_bytes_total{pod!=\"\"}[${RATE_WINDOW}]))")"

namespace_cpu_cores_used="$(query_prometheus "sum(${cpu_namespace_query})")"
namespace_memory_working_set_bytes="$(query_prometheus "sum(${memory_namespace_query})")"
namespace_network_receive_bytes_per_sec="$(query_prometheus "sum(rate(container_network_receive_bytes_total{namespace=\"${TARGET_NAMESPACE}\"}[${RATE_WINDOW}]))")"
namespace_network_transmit_bytes_per_sec="$(query_prometheus "sum(rate(container_network_transmit_bytes_total{namespace=\"${TARGET_NAMESPACE}\"}[${RATE_WINDOW}]))")"
namespace_running_pods="$(query_prometheus "sum(kube_pod_status_phase{namespace=\"${TARGET_NAMESPACE}\",phase=\"Running\"})")"

node_cpu_utilization_percent="$(query_prometheus "100 * (1 - avg(rate(node_cpu_seconds_total{mode=\"idle\"}[${RATE_WINDOW}])))")"
node_memory_utilization_percent="$(query_prometheus '100 * (1 - (sum(node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes)))')"

peer_cpu_cores_used="$(query_prometheus "${peer_cpu_total_query}")"
orderer_cpu_cores_used="$(query_prometheus "${orderer_cpu_total_query}")"
exp_app_cpu_cores_used="$(query_prometheus "${exp_app_cpu_total_query}")"
chaincode_cpu_cores_used="$(query_prometheus "${chaincode_cpu_total_query}")"

peer_memory_working_set_bytes="$(query_prometheus "${peer_memory_total_query}")"
orderer_memory_working_set_bytes="$(query_prometheus "${orderer_memory_total_query}")"
exp_app_memory_working_set_bytes="$(query_prometheus "${exp_app_memory_total_query}")"
chaincode_memory_working_set_bytes="$(query_prometheus "${chaincode_memory_total_query}")"

peer_cpu_per_pod="$(query_prometheus_map "${peer_cpu_per_pod_query}" "pod")"
orderer_cpu_per_pod="$(query_prometheus_map "${orderer_cpu_per_pod_query}" "pod")"
exp_app_cpu_per_pod="$(query_prometheus_map "${exp_app_cpu_per_pod_query}" "pod")"
chaincode_cpu_per_pod="$(query_prometheus_map "${chaincode_cpu_per_pod_query}" "pod")"

peer_memory_per_pod="$(query_prometheus_map "${peer_memory_per_pod_query}" "pod")"
orderer_memory_per_pod="$(query_prometheus_map "${orderer_memory_per_pod_query}" "pod")"
exp_app_memory_per_pod="$(query_prometheus_map "${exp_app_memory_per_pod_query}" "pod")"
chaincode_memory_per_pod="$(query_prometheus_map "${chaincode_memory_per_pod_query}" "pod")"

timeseries_json="null"
if [[ -n "${START_TS}" ]]; then
  peer_cpu_total_series="$(query_prometheus_range_values "${peer_cpu_total_query}")"
  orderer_cpu_total_series="$(query_prometheus_range_values "${orderer_cpu_total_query}")"
  exp_app_cpu_total_series="$(query_prometheus_range_values "${exp_app_cpu_total_query}")"
  chaincode_cpu_total_series="$(query_prometheus_range_values "${chaincode_cpu_total_query}")"

  peer_memory_total_series="$(query_prometheus_range_values "${peer_memory_total_query}")"
  orderer_memory_total_series="$(query_prometheus_range_values "${orderer_memory_total_query}")"
  exp_app_memory_total_series="$(query_prometheus_range_values "${exp_app_memory_total_query}")"
  chaincode_memory_total_series="$(query_prometheus_range_values "${chaincode_memory_total_query}")"

  peer_cpu_per_pod_series="$(query_prometheus_range_grouped "${peer_cpu_per_pod_query}" "pod")"
  orderer_cpu_per_pod_series="$(query_prometheus_range_grouped "${orderer_cpu_per_pod_query}" "pod")"
  exp_app_cpu_per_pod_series="$(query_prometheus_range_grouped "${exp_app_cpu_per_pod_query}" "pod")"
  chaincode_cpu_per_pod_series="$(query_prometheus_range_grouped "${chaincode_cpu_per_pod_query}" "pod")"

  peer_memory_per_pod_series="$(query_prometheus_range_grouped "${peer_memory_per_pod_query}" "pod")"
  orderer_memory_per_pod_series="$(query_prometheus_range_grouped "${orderer_memory_per_pod_query}" "pod")"
  exp_app_memory_per_pod_series="$(query_prometheus_range_grouped "${exp_app_memory_per_pod_query}" "pod")"
  chaincode_memory_per_pod_series="$(query_prometheus_range_grouped "${chaincode_memory_per_pod_query}" "pod")"

  timeseries_json="$(jq -n \
    --arg start_ts "${START_TS}" \
    --arg end_ts "${END_TS}" \
    --arg step "${STEP}" \
    --argjson peer_cpu_total_series "${peer_cpu_total_series}" \
    --argjson orderer_cpu_total_series "${orderer_cpu_total_series}" \
    --argjson exp_app_cpu_total_series "${exp_app_cpu_total_series}" \
    --argjson chaincode_cpu_total_series "${chaincode_cpu_total_series}" \
    --argjson peer_memory_total_series "${peer_memory_total_series}" \
    --argjson orderer_memory_total_series "${orderer_memory_total_series}" \
    --argjson exp_app_memory_total_series "${exp_app_memory_total_series}" \
    --argjson chaincode_memory_total_series "${chaincode_memory_total_series}" \
    --argjson peer_cpu_per_pod_series "${peer_cpu_per_pod_series}" \
    --argjson orderer_cpu_per_pod_series "${orderer_cpu_per_pod_series}" \
    --argjson exp_app_cpu_per_pod_series "${exp_app_cpu_per_pod_series}" \
    --argjson chaincode_cpu_per_pod_series "${chaincode_cpu_per_pod_series}" \
    --argjson peer_memory_per_pod_series "${peer_memory_per_pod_series}" \
    --argjson orderer_memory_per_pod_series "${orderer_memory_per_pod_series}" \
    --argjson exp_app_memory_per_pod_series "${exp_app_memory_per_pod_series}" \
    --argjson chaincode_memory_per_pod_series "${chaincode_memory_per_pod_series}" \
    '{
      window: {
        start_ts: ($start_ts | tonumber),
        end_ts: ($end_ts | tonumber),
        step: $step
      },
      component_totals: {
        cpu_cores_used: {
          peers: $peer_cpu_total_series,
          orderers: $orderer_cpu_total_series,
          exp_app: $exp_app_cpu_total_series,
          chaincodes: $chaincode_cpu_total_series
        },
        memory_working_set_bytes: {
          peers: $peer_memory_total_series,
          orderers: $orderer_memory_total_series,
          exp_app: $exp_app_memory_total_series,
          chaincodes: $chaincode_memory_total_series
        }
      },
      component_per_pod: {
        cpu_cores_used: {
          peers: $peer_cpu_per_pod_series,
          orderers: $orderer_cpu_per_pod_series,
          exp_app: $exp_app_cpu_per_pod_series,
          chaincodes: $chaincode_cpu_per_pod_series
        },
        memory_working_set_bytes: {
          peers: $peer_memory_per_pod_series,
          orderers: $orderer_memory_per_pod_series,
          exp_app: $exp_app_memory_per_pod_series,
          chaincodes: $chaincode_memory_per_pod_series
        }
      }
    }')"
fi

mkdir -p "$(dirname "${OUTPUT}")"

jq -n \
  --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg target_namespace "${TARGET_NAMESPACE}" \
  --arg monitoring_namespace "${MONITORING_NAMESPACE}" \
  --arg monitoring_release "${MONITORING_RELEASE_NAME}" \
  --argjson cluster_cpu_cores_used "${cluster_cpu_cores_used}" \
  --argjson cluster_memory_working_set_bytes "${cluster_memory_working_set_bytes}" \
  --argjson cluster_network_receive_bytes_per_sec "${cluster_network_receive_bytes_per_sec}" \
  --argjson cluster_network_transmit_bytes_per_sec "${cluster_network_transmit_bytes_per_sec}" \
  --argjson namespace_cpu_cores_used "${namespace_cpu_cores_used}" \
  --argjson namespace_memory_working_set_bytes "${namespace_memory_working_set_bytes}" \
  --argjson namespace_network_receive_bytes_per_sec "${namespace_network_receive_bytes_per_sec}" \
  --argjson namespace_network_transmit_bytes_per_sec "${namespace_network_transmit_bytes_per_sec}" \
  --argjson namespace_running_pods "${namespace_running_pods}" \
  --argjson node_cpu_utilization_percent "${node_cpu_utilization_percent}" \
  --argjson node_memory_utilization_percent "${node_memory_utilization_percent}" \
  --argjson peer_cpu_cores_used "${peer_cpu_cores_used}" \
  --argjson orderer_cpu_cores_used "${orderer_cpu_cores_used}" \
  --argjson exp_app_cpu_cores_used "${exp_app_cpu_cores_used}" \
  --argjson chaincode_cpu_cores_used "${chaincode_cpu_cores_used}" \
  --argjson peer_memory_working_set_bytes "${peer_memory_working_set_bytes}" \
  --argjson orderer_memory_working_set_bytes "${orderer_memory_working_set_bytes}" \
  --argjson exp_app_memory_working_set_bytes "${exp_app_memory_working_set_bytes}" \
  --argjson chaincode_memory_working_set_bytes "${chaincode_memory_working_set_bytes}" \
  --argjson peer_cpu_per_pod "${peer_cpu_per_pod}" \
  --argjson orderer_cpu_per_pod "${orderer_cpu_per_pod}" \
  --argjson exp_app_cpu_per_pod "${exp_app_cpu_per_pod}" \
  --argjson chaincode_cpu_per_pod "${chaincode_cpu_per_pod}" \
  --argjson peer_memory_per_pod "${peer_memory_per_pod}" \
  --argjson orderer_memory_per_pod "${orderer_memory_per_pod}" \
  --argjson exp_app_memory_per_pod "${exp_app_memory_per_pod}" \
  --argjson chaincode_memory_per_pod "${chaincode_memory_per_pod}" \
  --argjson timeseries "${timeseries_json}" \
  '{
    timestamp: $timestamp,
    target_namespace: $target_namespace,
    monitoring: {
      namespace: $monitoring_namespace,
      release: $monitoring_release
    },
    metrics: {
      cluster_cpu_cores_used: $cluster_cpu_cores_used,
      cluster_memory_working_set_bytes: $cluster_memory_working_set_bytes,
      cluster_network_receive_bytes_per_sec: $cluster_network_receive_bytes_per_sec,
      cluster_network_transmit_bytes_per_sec: $cluster_network_transmit_bytes_per_sec,
      namespace_cpu_cores_used: $namespace_cpu_cores_used,
      namespace_memory_working_set_bytes: $namespace_memory_working_set_bytes,
      namespace_network_receive_bytes_per_sec: $namespace_network_receive_bytes_per_sec,
      namespace_network_transmit_bytes_per_sec: $namespace_network_transmit_bytes_per_sec,
      namespace_running_pods: $namespace_running_pods,
      node_cpu_utilization_percent: $node_cpu_utilization_percent,
      node_memory_utilization_percent: $node_memory_utilization_percent,
      component_totals: {
        cpu_cores_used: {
          peers: $peer_cpu_cores_used,
          orderers: $orderer_cpu_cores_used,
          exp_app: $exp_app_cpu_cores_used,
          chaincodes: $chaincode_cpu_cores_used
        },
        memory_working_set_bytes: {
          peers: $peer_memory_working_set_bytes,
          orderers: $orderer_memory_working_set_bytes,
          exp_app: $exp_app_memory_working_set_bytes,
          chaincodes: $chaincode_memory_working_set_bytes
        }
      },
      component_per_pod: {
        cpu_cores_used: {
          peers: $peer_cpu_per_pod,
          orderers: $orderer_cpu_per_pod,
          exp_app: $exp_app_cpu_per_pod,
          chaincodes: $chaincode_cpu_per_pod
        },
        memory_working_set_bytes: {
          peers: $peer_memory_per_pod,
          orderers: $orderer_memory_per_pod,
          exp_app: $exp_app_memory_per_pod,
          chaincodes: $chaincode_memory_per_pod
        }
      }
    },
    timeseries: $timeseries
  }' > "${OUTPUT}"

echo "Cluster resource metrics saved: ${OUTPUT}"
