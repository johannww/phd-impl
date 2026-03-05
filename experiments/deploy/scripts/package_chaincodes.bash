CHAINCODE_PACKAGE_CONFIGMAP="${CHAINCODE_PACKAGE_CONFIGMAP:-${CHAINCODE_RELEASE_NAME}-packages}"
PEER_ORGS="${PEER_ORGS:-org1}"
CARBON_CC_VERSION="${CARBON_CC_VERSION:-1.0}"
INTEROP_CC_VERSION="${INTEROP_CC_VERSION:-1.0}"
CARBON_CC_PORT="${CARBON_CC_PORT:-7055}"
INTEROP_CC_PORT="${INTEROP_CC_PORT:-7055}"

PKG_DIR="${SCRIPT_DIR}/chaincode-packages"
mkdir -p "${PKG_DIR}"
rm -f "${PKG_DIR}"/*.tar.gz

CC_SET_ARGS=()

split_image_ref() {
  local image_ref="$1"
  if [[ "${image_ref}" == *:* ]]; then
    IMAGE_REPO="${image_ref%:*}"
    IMAGE_TAG="${image_ref##*:}"
  else
    IMAGE_REPO="${image_ref}"
    IMAGE_TAG="latest"
  fi
}

package_chaincode() {
  local cc_name="$1"
  local cc_version="$2"
  local cc_port="$3"
  local cc_index="$4"
  local cc_image="$5"
  local cc_label="${cc_name}_${cc_version}"
  split_image_ref "${cc_image}"
  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].name=${cc_name}")
  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].version=${cc_version}")
  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].packageLabel=${cc_label}")
  CC_SET_ARGS+=(--set "chaincodes[${cc_index}].servicePort=${cc_port}")
  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].image.repository=${IMAGE_REPO}")
  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].image.tag=${IMAGE_TAG}")
  CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].image.pullPolicy=IfNotPresent")
  for org in ${PEER_ORGS}; do
    local cc_file="${PKG_DIR}/${cc_name}_${org}.tar.gz"
    local tmpdir
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
    local digest
    digest="$(sha256sum "${cc_file}" | awk '{print $1}')"
    CC_SET_ARGS+=(--set-string "chaincodes[${cc_index}].packageIds.${org}=${cc_label}:${digest}")
    rm -rf "${tmpdir}"
  done
}

echo "Packaging chaincode service archives..."
package_chaincode "carbon" "${CARBON_CC_VERSION}" "${CARBON_CC_PORT}" "0" "${CARBON_CC_IMAGE}"
package_chaincode "interop" "${INTEROP_CC_VERSION}" "${INTEROP_CC_PORT}" "1" "${INTEROP_CC_IMAGE}"

echo "Uploading chaincode packages configmap ${CHAINCODE_PACKAGE_CONFIGMAP}..."
kubectl -n "${NAMESPACE}" create configmap "${CHAINCODE_PACKAGE_CONFIGMAP}" \
  --from-file="${PKG_DIR}" \
  --dry-run=client -o yaml | kubectl -n "${NAMESPACE}" apply -f -
