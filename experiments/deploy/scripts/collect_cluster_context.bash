#!/usr/bin/env bash
set -euo pipefail

OUTPUT=""
RESOURCE_GROUP="${RESOURCE_GROUP:-carbon}"
CLUSTER_NAME="${CLUSTER_NAME:-carbon-aks}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      OUTPUT="$2"
      shift 2
      ;;
    --resource-group)
      RESOURCE_GROUP="$2"
      shift 2
      ;;
    --cluster-name)
      CLUSTER_NAME="$2"
      shift 2
      ;;
    *)
      echo "Usage: $0 --output <file> [--resource-group group] [--cluster-name name]"
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

NODES_JSON="$(kubectl get nodes -o json)"

is_aks="false"
if jq -e 'all(.items[]?; ((.spec.providerID // "") | startswith("azure:///")))' >/dev/null 2>&1 <<< "${NODES_JSON}"; then
  is_aks="true"
fi

cluster_summary_json="$(jq -c '
  def cpu_to_m(v):
    if (v | type) == "string" then
      if v | endswith("m") then (v[0:-1] | tonumber) else ((v | tonumber) * 1000) end
    else v end;

  def mem_to_ki(v):
    if (v | type) == "string" then
      if v | endswith("Ki") then (v[0:-2] | tonumber)
      elif v | endswith("Mi") then ((v[0:-2] | tonumber) * 1024)
      elif v | endswith("Gi") then ((v[0:-2] | tonumber) * 1024 * 1024)
      else (v | tonumber) end
    else v end;

  def is_ready: any(.status.conditions[]?; .type == "Ready" and .status == "True");

  {
    node_count: (.items | length),
    ready_node_count: ([.items[] | select(is_ready)] | length),
    nodes: [
      .items[]
      | {
          name: .metadata.name,
          ready: is_ready,
          provider_id: (.spec.providerID // null),
          instance_type: (.metadata.labels["node.kubernetes.io/instance-type"] // null),
          agentpool: (.metadata.labels["kubernetes.azure.com/agentpool"] // .metadata.labels["agentpool"] // null),
          capacity: {
            cpu_millicores: (.status.capacity.cpu | cpu_to_m(.)),
            memory_kib: (.status.capacity.memory | mem_to_ki(.)),
            pods: (.status.capacity.pods | tonumber)
          },
          allocatable: {
            cpu_millicores: (.status.allocatable.cpu | cpu_to_m(.)),
            memory_kib: (.status.allocatable.memory | mem_to_ki(.)),
            pods: (.status.allocatable.pods | tonumber)
          }
        }
    ],
    totals: {
      capacity: {
        cpu_millicores: ([.items[] | select(is_ready) | .status.capacity.cpu | cpu_to_m(.)] | add // 0),
        memory_kib: ([.items[] | select(is_ready) | .status.capacity.memory | mem_to_ki(.)] | add // 0),
        pods: ([.items[] | select(is_ready) | .status.capacity.pods | tonumber] | add // 0)
      },
      allocatable: {
        cpu_millicores: ([.items[] | select(is_ready) | .status.allocatable.cpu | cpu_to_m(.)] | add // 0),
        memory_kib: ([.items[] | select(is_ready) | .status.allocatable.memory | mem_to_ki(.)] | add // 0),
        pods: ([.items[] | select(is_ready) | .status.allocatable.pods | tonumber] | add // 0)
      }
    }
  }
' <<< "${NODES_JSON}")"

aks_cluster_json='null'
aks_nodepools_json='[]'
if [[ "${is_aks}" == "true" ]] && command -v az >/dev/null 2>&1 && az account show >/dev/null 2>&1; then
  aks_cluster_json="$(az aks show \
    --resource-group "${RESOURCE_GROUP}" \
    --name "${CLUSTER_NAME}" \
    --query '{name:name,location:location,kubernetes_version:kubernetesVersion,power_state:powerState.code,node_resource_group:nodeResourceGroup,dns_prefix:dnsPrefix}' \
    -o json 2>/dev/null || printf 'null')"
  aks_nodepools_json="$(az aks nodepool list \
    --resource-group "${RESOURCE_GROUP}" \
    --cluster-name "${CLUSTER_NAME}" \
    --query '[].{name:name,mode:mode,count:count,vm_size:vmSize,os_type:osType,max_pods:maxPods,enable_auto_scaling:enableAutoScaling,min_count:minCount,max_count:maxCount,provisioning_state:provisioningState,power_state:powerState.code}' \
    -o json 2>/dev/null || printf '[]')"
fi

mkdir -p "$(dirname "${OUTPUT}")"

jq -n \
  --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg provider "$(if [[ "${is_aks}" == "true" ]]; then echo aks; else echo kubernetes; fi)" \
  --arg resource_group "${RESOURCE_GROUP}" \
  --arg cluster_name "${CLUSTER_NAME}" \
  --argjson cluster_summary "${cluster_summary_json}" \
  --argjson aks_cluster "${aks_cluster_json}" \
  --argjson aks_nodepools "${aks_nodepools_json}" \
  '{
    timestamp: $timestamp,
    provider: $provider,
    kubernetes: $cluster_summary,
    aks: (if $provider == "aks" then {
      resource_group: $resource_group,
      cluster_name: $cluster_name,
      cluster: $aks_cluster,
      nodepools: $aks_nodepools
    } else null end)
  }' > "${OUTPUT}"

echo "Cluster context saved: ${OUTPUT}"
