#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

OUTPUT=""
TARGET_NAMESPACE="${TARGET_NAMESPACE:-fabric-experiments}"
MONITORING_NAMESPACE="${MONITORING_NAMESPACE:-monitoring}"
MONITORING_RELEASE_NAME="${MONITORING_RELEASE_NAME:-monitoring}"
START_TS=""
END_TS=""
STEP="15s"
RATE_WINDOW="${RATE_WINDOW:-5m}"
ENABLE_TEE_AUCTION_METRICS="${ENABLE_TEE_AUCTION_METRICS:-true}"
TEE_RESOURCE_GROUP="${TEE_RESOURCE_GROUP:-${RESOURCE_GROUP:-carbon}}"
TEE_CONTAINER_NAME="${TEE_CONTAINER_NAME:-carbon-auction-container}"
TEE_METRICS_URL="${TEE_METRICS_URL:-}"
TEE_METRICS_SCHEME="${TEE_METRICS_SCHEME:-https}"
TEE_METRICS_PORT="${TEE_METRICS_PORT:-8080}"
TEE_METRICS_PATH="${TEE_METRICS_PATH:-/metrics}"
TEE_METRICS_INSECURE="${TEE_METRICS_INSECURE:-true}"
TEE_AZURE_INTERVAL="${TEE_AZURE_INTERVAL:-PT1M}"

TIMESERIES_FILE="$(mktemp)"
FABRIC_STORAGE_FILE="$(mktemp)"
TEE_AZURE_SERIES_FILE="$(mktemp)"
trap 'rm -f "${TIMESERIES_FILE}" "${FABRIC_STORAGE_FILE}" "${TEE_AZURE_SERIES_FILE}"' EXIT

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

epoch_to_iso8601() {
  date -u -d "@$1" +%Y-%m-%dT%H:%M:%SZ
}

query_tee_container_context() {
  if [[ "${ENABLE_TEE_AUCTION_METRICS}" != "true" ]]; then
    echo '{}'
    return 0
  fi

  if ! command -v az >/dev/null 2>&1; then
    echo '{}'
    return 0
  fi

  if ! az account show >/dev/null 2>&1; then
    echo '{}'
    return 0
  fi

  az container show \
    --resource-group "${TEE_RESOURCE_GROUP}" \
    --name "${TEE_CONTAINER_NAME}" \
    --query '{resource_id:id,ip:ipAddress.ip,location:location,name:name}' \
    -o json 2>/dev/null || echo '{}'
}

query_tee_azure_metrics_snapshot() {
  local resource_id="$1"

  if [[ -z "${resource_id}" ]]; then
    echo 'null'
    return 0
  fi

  local end_iso start_iso raw now_epoch
  now_epoch="$(date +%s)"
  end_iso="$(epoch_to_iso8601 "${now_epoch}")"
  start_iso="$(epoch_to_iso8601 "$((now_epoch - 600))")"

  raw="$(az monitor metrics list \
    --resource "${resource_id}" \
    --metrics CPUUsage MemoryUsage \
    --aggregation Maximum \
    --interval "${TEE_AZURE_INTERVAL}" \
    --start-time "${start_iso}" \
    --end-time "${end_iso}" \
    -o json 2>/dev/null || true)"

  if [[ -z "${raw}" ]]; then
    echo 'null'
    return 0
  fi

  jq -c '
    def max_sample($names):
      ([.value[]? as $metric | select(($names | index($metric.name.value)) != null) | $metric.timeseries[]?.data[]? | .maximum | select(. != null)] | max) // null;

    {
      cpu_usage_burst_max_millicores: max_sample(["CPUUsage", "CpuUsage"]),
      memory_usage_burst_max_bytes: max_sample(["MemoryUsage"])
    }
  ' <<< "${raw}"
}

query_tee_azure_metrics_series() {
  local resource_id="$1"

  if [[ -z "${resource_id}" || -z "${START_TS}" ]]; then
    echo 'null'
    return 0
  fi

  local raw start_iso end_iso
  start_iso="$(epoch_to_iso8601 "${START_TS}")"
  end_iso="$(epoch_to_iso8601 "${END_TS}")"

  raw="$(az monitor metrics list \
    --resource "${resource_id}" \
    --metrics CPUUsage MemoryUsage \
    --aggregation Maximum \
    --interval "${TEE_AZURE_INTERVAL}" \
    --start-time "${start_iso}" \
    --end-time "${end_iso}" \
    -o json 2>/dev/null || true)"

  if [[ -z "${raw}" ]]; then
    echo 'null'
    return 0
  fi

  jq -c '
    def series($names):
      [
        .value[]? as $metric
        | select(($names | index($metric.name.value)) != null)
        | $metric.timeseries[]?.data[]?
        | select(.maximum != null)
        | {ts: (.timeStamp | fromdateiso8601), value: .maximum}
      ];

    {
      cpu_usage_burst_max_millicores: series(["CPUUsage", "CpuUsage"]),
      memory_usage_burst_max_bytes: series(["MemoryUsage"])
    }
  ' <<< "${raw}"
}

query_tee_app_metrics() {
  local metrics_url="$1"

  if [[ -z "${metrics_url}" ]]; then
    echo 'null'
    return 0
  fi

  local curl_args=(-fsS)
  if [[ "${TEE_METRICS_INSECURE}" == "true" ]]; then
    curl_args+=(-k)
  fi

  local raw
  raw="$(curl "${curl_args[@]}" "${metrics_url}" 2>/dev/null || true)"
  if [[ -z "${raw}" ]]; then
    echo 'null'
    return 0
  fi

  jq -Rn '
    def parse_labels($text):
      if ($text == null or $text == "") then {}
      else
        reduce ($text | split(",")[]) as $pair ({};
          if ($pair | length) == 0 then .
          else ($pair | capture("(?<k>[^=]+)=\"(?<v>[^\"]*)\"")) as $m | .[$m.k] = $m.v
          end
        )
      end;

    def add_to_map($map; $key; $value):
      $map + {($key): (($map[$key] // 0) + $value)};

    def route_key($labels):
      ($labels.method // "unknown") + " " + ($labels.route // "unknown") + " " + ($labels.status // "unknown");

    reduce inputs as $line (
      {
        http_requests_total: {total: 0, by_route: {}},
        http_request_duration_seconds: {sum: 0, count: 0, by_route_sum: {}, by_route_count: {}},
        http_in_flight_requests: 0,
        auction_requests_total: {},
        auction_run_duration_seconds: {sum: 0, count: 0},
        report_deserialize_total: {}
      };
      if ($line | test("^tee_auction_http_requests_total(\\{|\\s)")) then
        ($line | capture("^tee_auction_http_requests_total(?:\\{(?<labels>[^}]*)\\})?\\s+(?<value>[-+0-9.eE]+)$")) as $m |
        ($m.value | tonumber) as $value |
        (parse_labels($m.labels // "")) as $labels |
        .http_requests_total.total += $value |
        .http_requests_total.by_route = add_to_map(.http_requests_total.by_route; route_key($labels); $value)
      elif ($line | test("^tee_auction_http_request_duration_seconds_sum(\\{|\\s)")) then
        ($line | capture("^tee_auction_http_request_duration_seconds_sum(?:\\{(?<labels>[^}]*)\\})?\\s+(?<value>[-+0-9.eE]+)$")) as $m |
        ($m.value | tonumber) as $value |
        (parse_labels($m.labels // "")) as $labels |
        .http_request_duration_seconds.sum += $value |
        .http_request_duration_seconds.by_route_sum = add_to_map(.http_request_duration_seconds.by_route_sum; route_key($labels); $value)
      elif ($line | test("^tee_auction_http_request_duration_seconds_count(\\{|\\s)")) then
        ($line | capture("^tee_auction_http_request_duration_seconds_count(?:\\{(?<labels>[^}]*)\\})?\\s+(?<value>[-+0-9.eE]+)$")) as $m |
        ($m.value | tonumber) as $value |
        (parse_labels($m.labels // "")) as $labels |
        .http_request_duration_seconds.count += $value |
        .http_request_duration_seconds.by_route_count = add_to_map(.http_request_duration_seconds.by_route_count; route_key($labels); $value)
      elif ($line | test("^tee_auction_http_in_flight_requests\\s")) then
        ($line | capture("^tee_auction_http_in_flight_requests\\s+(?<value>[-+0-9.eE]+)$").value | tonumber) as $value |
        .http_in_flight_requests = $value
      elif ($line | test("^tee_auction_auction_requests_total(\\{|\\s)")) then
        ($line | capture("^tee_auction_auction_requests_total(?:\\{(?<labels>[^}]*)\\})?\\s+(?<value>[-+0-9.eE]+)$")) as $m |
        ($m.value | tonumber) as $value |
        (parse_labels($m.labels // "")) as $labels |
        .auction_requests_total = add_to_map(.auction_requests_total; ($labels.result // "unknown"); $value)
      elif ($line | test("^tee_auction_auction_run_duration_seconds_sum\\s")) then
        ($line | capture("^tee_auction_auction_run_duration_seconds_sum\\s+(?<value>[-+0-9.eE]+)$").value | tonumber) as $value |
        .auction_run_duration_seconds.sum = $value
      elif ($line | test("^tee_auction_auction_run_duration_seconds_count\\s")) then
        ($line | capture("^tee_auction_auction_run_duration_seconds_count\\s+(?<value>[-+0-9.eE]+)$").value | tonumber) as $value |
        .auction_run_duration_seconds.count = $value
      elif ($line | test("^tee_auction_report_deserialize_total(\\{|\\s)")) then
        ($line | capture("^tee_auction_report_deserialize_total(?:\\{(?<labels>[^}]*)\\})?\\s+(?<value>[-+0-9.eE]+)$")) as $m |
        ($m.value | tonumber) as $value |
        (parse_labels($m.labels // "")) as $labels |
        .report_deserialize_total = add_to_map(.report_deserialize_total; ($labels.result // "unknown"); $value)
      else
        .
      end
    )
  ' <<< "${raw}"
}

peer_pod_regex=".*-peer-.*"
orderer_pod_regex=".*-orderer-.*"
exp_app_pod_regex="^exp-app-.*"
chaincode_pod_regex="^(carbon|interop)-.*"

TEE_CONTEXT_JSON="$(query_tee_container_context)"
TEE_RESOURCE_ID="$(jq -r '.resource_id // empty' <<< "${TEE_CONTEXT_JSON}")"
TEE_IP="$(jq -r '.ip // empty' <<< "${TEE_CONTEXT_JSON}")"

if [[ -z "${TEE_METRICS_URL}" && -n "${TEE_IP}" ]]; then
  TEE_METRICS_URL="${TEE_METRICS_SCHEME}://${TEE_IP}:${TEE_METRICS_PORT}${TEE_METRICS_PATH}"
fi

TEE_AZURE_METRICS_SNAPSHOT="$(query_tee_azure_metrics_snapshot "${TEE_RESOURCE_ID}")"
TEE_APP_METRICS_SNAPSHOT="$(query_tee_app_metrics "${TEE_METRICS_URL}")"

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

printf 'null' > "${TIMESERIES_FILE}"
printf 'null' > "${TEE_AZURE_SERIES_FILE}"
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

  jq -n \
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
  }' > "${TIMESERIES_FILE}"

  query_tee_azure_metrics_series "${TEE_RESOURCE_ID}" > "${TEE_AZURE_SERIES_FILE}"
fi

"${SCRIPT_DIR}/collect_fabric_storage_metrics.bash" \
  --output "${FABRIC_STORAGE_FILE}" \
  --target-namespace "${TARGET_NAMESPACE}" >/dev/null

mkdir -p "$(dirname "${OUTPUT}")"

jq -n \
  --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg target_namespace "${TARGET_NAMESPACE}" \
  --arg monitoring_namespace "${MONITORING_NAMESPACE}" \
  --arg monitoring_release "${MONITORING_RELEASE_NAME}" \
  --arg tee_resource_group "${TEE_RESOURCE_GROUP}" \
  --arg tee_container_name "${TEE_CONTAINER_NAME}" \
  --arg tee_resource_id "${TEE_RESOURCE_ID}" \
  --arg tee_ip "${TEE_IP}" \
  --arg tee_metrics_url "${TEE_METRICS_URL}" \
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
  --argjson tee_azure_metrics_snapshot "${TEE_AZURE_METRICS_SNAPSHOT}" \
  --argjson tee_app_metrics_snapshot "${TEE_APP_METRICS_SNAPSHOT}" \
  --slurpfile fabric_storage "${FABRIC_STORAGE_FILE}" \
  --slurpfile tee_azure_series "${TEE_AZURE_SERIES_FILE}" \
  --slurpfile timeseries "${TIMESERIES_FILE}" \
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
      },
      external_components: {
        tee_auction: {
          azure_monitor: ($tee_azure_metrics_snapshot // null),
          app_metrics: ($tee_app_metrics_snapshot // null)
        }
      },
      fabric_storage: (($fabric_storage[0].metrics) // {
        peers: {},
        orderers: {},
        totals: {
          peers: {pod_count: 0, ledger_root_bytes: 0},
          orderers: {pod_count: 0, ledger_root_bytes: 0},
          all: {pod_count: 0, ledger_root_bytes: 0}
        }
      })
    },
    external_components_inventory: {
      tee_auction: {
        resource_group: $tee_resource_group,
        container_name: $tee_container_name,
        resource_id: $tee_resource_id,
        ip: $tee_ip,
        metrics_url: $tee_metrics_url
      }
    },
    fabric_storage_inventory: (($fabric_storage[0].inventory) // {peers: {}, orderers: {}}),
    timeseries: (($timeseries[0] // null) | if . == null then null else . + {
      external_components: {
        tee_auction: {
          azure_monitor: (($tee_azure_series[0]) // null)
        }
      }
    } end)
  }' > "${OUTPUT}"

echo "Cluster resource metrics saved: ${OUTPUT}"
