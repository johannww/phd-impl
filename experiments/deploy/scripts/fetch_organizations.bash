#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
VALUES_FILE="${VALUES_FILE:-${SCRIPT_DIR}/../helm/values.yaml}"
LOCAL_DIR="${LOCAL_DIR:-${SCRIPT_DIR}/../vars/organizations}"

# Pick the first peer org's peer-0 pod via label selector
PEER_ORG="$(yq e '.network.organizations[] | select(.peers) | .name' "${VALUES_FILE}" | head -1)"
POD=""
while [[ -z "${POD}" ]]; do
  echo "Waiting for peer pod to be ready..."
  POD="$(kubectl -n "${NAMESPACE}" get pod \
    -l "app.kubernetes.io/component=peer,app.kubernetes.io/org=${PEER_ORG},app.kubernetes.io/node-id=peer-0" \
    -o jsonpath='{.items[0].metadata.name}')"
  sleep 2
done

mkdir -p "${LOCAL_DIR}"
echo "Copying /organizations from pod ${POD} (${NAMESPACE}) to ${LOCAL_DIR}..."
kubectl -n "${NAMESPACE}" cp "${POD}:/organizations/." "${LOCAL_DIR}/" -c peer
echo "Done."
