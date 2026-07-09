#!/usr/bin/env bash
set -euo pipefail

OUTPUT=""
TARGET_NAMESPACE="${TARGET_NAMESPACE:-fabric-experiments}"

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
		*)
			echo "Usage: $0 --output <file> [--target-namespace ns]"
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

collect_pod_snapshot() {
	local pod_name="$1"
	local component="$2"
	local org_name="$3"

	kubectl exec -n "${TARGET_NAMESPACE}" -c "${component}" "${pod_name}" -- /bin/sh -lc '
		set -eu

		ledger_root=""
		for candidate in "${CORE_PEER_FILESYSTEMPATH:-}" "${ORDERER_FILELEDGER_LOCATION:-}" /var/hyperledger/production; do
			if [ -n "${candidate}" ] && [ -e "${candidate}" ]; then
				ledger_root="${candidate}"
				break
			fi
		done

		if [ -z "${ledger_root}" ]; then
			echo "Error: could not resolve Fabric ledger root" >&2
			exit 1
		fi

		root_bytes=$(du -sb "${ledger_root}" | awk "{print \$1}")
		printf "ledger_root\t%s\n" "${ledger_root}"
		printf "ledger_root_bytes\t%s\n" "${root_bytes}"

		df -B1 "${ledger_root}" | awk "NR==2 {printf \"filesystem\\t%s\\t%s\\t%s\\t%s\\t%s\\n\", \$1, \$2, \$3, \$4, \$6}"

		find "${ledger_root}" -mindepth 1 -maxdepth 2 -type d | sort | while IFS= read -r dir; do
			rel="${dir#${ledger_root}/}"
			bytes=$(du -sb "${dir}" | awk "{print \$1}")
			printf "dir\t%s\t%s\n" "${rel}" "${bytes}"
		done
	' | jq -Rn \
		--arg pod_name "${pod_name}" \
		--arg component "${component}" \
		--arg org_name "${org_name}" '
			reduce inputs as $line (
				{
					pod: $pod_name,
					component: $component,
					org: $org_name,
					ledger_root: null,
					ledger_root_bytes: 0,
					filesystem: {},
					subdirs: {}
				};
				($line | split("\t")) as $parts |
				if ($parts[0] == "ledger_root") and (($parts | length) >= 2) then
					.ledger_root = $parts[1]
				elif ($parts[0] == "ledger_root_bytes") and (($parts | length) >= 2) then
					.ledger_root_bytes = (($parts[1] | tonumber?) // 0)
				elif ($parts[0] == "filesystem") and (($parts | length) >= 6) then
					.filesystem = {
						source: $parts[1],
						bytes_total: (($parts[2] | tonumber?) // 0),
						bytes_used: (($parts[3] | tonumber?) // 0),
						bytes_available: (($parts[4] | tonumber?) // 0),
						mount_point: $parts[5]
					}
				elif ($parts[0] == "dir") and (($parts | length) >= 3) then
					.subdirs[$parts[1]] = (($parts[2] | tonumber?) // 0)
				else
					.
				end
			)
		'
}

collect_component_snapshots() {
	local component="$1"
	local output_file="$2"
	local selector="app.kubernetes.io/component=${component}"

	: > "${output_file}"

	while IFS= read -r pod_json; do
		[[ -z "${pod_json}" ]] && continue

		local pod_name org_name
		pod_name="$(jq -r '.name' <<< "${pod_json}")"
		org_name="$(jq -r '.org' <<< "${pod_json}")"

		collect_pod_snapshot "${pod_name}" "${component}" "${org_name}" >> "${output_file}"
	done < <(
		kubectl get pods -n "${TARGET_NAMESPACE}" -l "${selector}" -o json |
			jq -rc '.items[] | {name: .metadata.name, org: (.metadata.labels["app.kubernetes.io/org"] // "")}'
	)
}

PEER_SNAPSHOTS_FILE="$(mktemp)"
ORDERER_SNAPSHOTS_FILE="$(mktemp)"
trap 'rm -f "${PEER_SNAPSHOTS_FILE}" "${ORDERER_SNAPSHOTS_FILE}"' EXIT

collect_component_snapshots "peer" "${PEER_SNAPSHOTS_FILE}"
collect_component_snapshots "orderer" "${ORDERER_SNAPSHOTS_FILE}"

mkdir -p "$(dirname "${OUTPUT}")"

jq -n \
	--arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
	--arg target_namespace "${TARGET_NAMESPACE}" \
	--slurpfile peers "${PEER_SNAPSHOTS_FILE}" \
	--slurpfile orderers "${ORDERER_SNAPSHOTS_FILE}" '
		def inventory_map($items):
			reduce $items[] as $item ({};
				.[$item.pod] = {
					component: $item.component,
					org: $item.org,
					ledger_root: $item.ledger_root,
					filesystem: {
						source: ($item.filesystem.source // null),
						mount_point: ($item.filesystem.mount_point // null)
					}
				}
			);

		def metrics_map($items):
			reduce $items[] as $item ({};
				.[$item.pod] = {
					ledger_root_bytes: ($item.ledger_root_bytes // 0),
					subdirs: ($item.subdirs // {}),
					filesystem: {
						bytes_total: ($item.filesystem.bytes_total // 0),
						bytes_used: ($item.filesystem.bytes_used // 0),
						bytes_available: ($item.filesystem.bytes_available // 0)
					}
				}
			);

		def total_ledger_bytes($items):
			reduce $items[] as $item (0; . + ($item.ledger_root_bytes // 0));

		{
			timestamp: $timestamp,
			target_namespace: $target_namespace,
			inventory: {
				peers: inventory_map($peers),
				orderers: inventory_map($orderers)
			},
			metrics: {
				peers: metrics_map($peers),
				orderers: metrics_map($orderers),
				totals: {
					peers: {
						pod_count: ($peers | length),
						ledger_root_bytes: total_ledger_bytes($peers)
					},
					orderers: {
						pod_count: ($orderers | length),
						ledger_root_bytes: total_ledger_bytes($orderers)
					},
					all: {
						pod_count: (($peers | length) + ($orderers | length)),
						ledger_root_bytes: (total_ledger_bytes($peers) + total_ledger_bytes($orderers))
					}
				}
			}
		}
	' > "${OUTPUT}"

echo "Fabric storage metrics saved: ${OUTPUT}"
