#!/usr/bin/env bash
# Runs the SICAR mock service Hurl test suite (test_sicar.hurl) against every
# property in the locally fetched sicar.json artifact.
#
# Usage:
#   ./test_sicar.bash
#
# Environment variables:
#   NAMESPACE   Kubernetes namespace (default: fabric-experiments)
#   CA_CERT     Path to the SICAR CA cert (default: vars/sicar/ca.crt)
#   SICAR_JSON  Path to the canonical sicar.json (default: vars/organizations/sicar/sicar.json)
#   SICAR_PORT  NodePort of the sicar-mock service (default: 30443)
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
VARS_DIR="${VARS_DIR:-${SCRIPT_DIR}/../vars}"
SICAR_JSON="${SICAR_JSON:-${VARS_DIR}/organizations/sicar/sicar.json}"
CA_CERT="${CA_CERT:-${VARS_DIR}/sicar/ca.crt}"
NAMESPACE="${NAMESPACE:-fabric-experiments}"
SICAR_PORT="${SICAR_PORT:-30443}"

if [[ ! -f "${SICAR_JSON}" ]]; then
  echo "ERROR: sicar.json not found at ${SICAR_JSON}" >&2
  echo "Run fetch_organizations.bash first, or set SICAR_JSON." >&2
  exit 1
fi

# Fetch the CA cert from the cluster if not already present locally.
if [[ ! -f "${CA_CERT}" ]]; then
  echo "CA cert not found at ${CA_CERT}, fetching from cluster..."
  kubectl get configmap sicar-mock-ca \
    -n "${NAMESPACE}" \
    -o jsonpath='{.data.ca\.crt}' > "${CA_CERT}"
  echo "Saved to ${CA_CERT}"
fi

MINIKUBE_IP="$(minikube ip)"
BASE_URL="https://${MINIKUBE_IP}:${SICAR_PORT}"

IDS="$(jq -r 'keys[]' "${SICAR_JSON}")"
TOTAL="$(echo "${IDS}" | wc -l)"

echo "Testing ${TOTAL} imoveis against ${BASE_URL} ..."

FAILED=0
while IFS= read -r codigo_imovel; do
  hurl --cacert "${CA_CERT}" \
       -k \
       --no-output \
       --variable "base_url=${BASE_URL}" \
       --variable "codigo_imovel=${codigo_imovel}" \
       "${SCRIPT_DIR}/test_sicar.hurl" \
  || FAILED=$((FAILED + 1))
done <<< "${IDS}"

if [[ "${FAILED}" -gt 0 ]]; then
  echo "FAILED: ${FAILED}/${TOTAL} imoveis had errors." >&2
  exit 1
fi

echo "All ${TOTAL} imoveis passed."
