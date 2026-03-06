VALUES_FILE="${VALUES_FILE:-${SCRIPT_DIR}/../helm/values.yaml}"
CHAINCODE_PACKAGE_CONFIGMAP="${CHAINCODE_PACKAGE_CONFIGMAP:-${CHAINCODE_RELEASE_NAME}-packages}"

PKG_DIR="${SCRIPT_DIR}/../vars/chaincode-packages"
mkdir -p "${PKG_DIR}"
rm -f "${PKG_DIR}"/*.tar.gz

CC_SET_ARGS=()

# Read peer orgs (orgs that have a 'peers' field defined)
mapfile -t PEER_ORGS < <(yq e '.chaincodeService.network.organizations[] | select(.peers) | .name' "${VALUES_FILE}")

echo "Packaging chaincode service archives..."
cc_count=$(yq e '.chaincodeService.chaincodes | length' "${VALUES_FILE}")
for ((cc_index=0; cc_index<cc_count; cc_index++)); do
  cc_name=$(yq e ".chaincodeService.chaincodes[${cc_index}].name" "${VALUES_FILE}")
  cc_version=$(yq e ".chaincodeService.chaincodes[${cc_index}].version" "${VALUES_FILE}")
  cc_port=$(yq e ".chaincodeService.chaincodes[${cc_index}].servicePort" "${VALUES_FILE}")
  cc_label="${cc_name}_${cc_version}"

  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].packageLabel=${cc_label}")

  for org in "${PEER_ORGS[@]}"; do
    cc_file="${PKG_DIR}/${cc_name}_${org}.tar.gz"
    tmpdir="$(mktemp -d)"
    cat > "${tmpdir}/connection.json" <<EOF
{
  "address": "${cc_name}-${org}:${cc_port}",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF
    cat > "${tmpdir}/metadata.json" <<EOF
{
  "type": "ccaas",
  "label": "${cc_label}"
}
EOF
    (cd "${tmpdir}" && tar cfz code.tar.gz connection.json && tar cfz "${cc_file}" metadata.json code.tar.gz)
    digest="$(sha256sum "${cc_file}" | awk '{print $1}')"
    CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].packageIds.${org}=${cc_label}:${digest}")
    rm -rf "${tmpdir}"
  done
done

echo "Uploading chaincode packages configmap ${CHAINCODE_PACKAGE_CONFIGMAP}..."
kubectl -n "${NAMESPACE}" create configmap "${CHAINCODE_PACKAGE_CONFIGMAP}" \
  --from-file="${PKG_DIR}" \
  --dry-run=client -o yaml | kubectl -n "${NAMESPACE}" apply -f -
