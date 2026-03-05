if ! command -v peer &>/dev/null; then
  FABRIC_SRC_DIR="$(mktemp -d)"
  echo "Cloning Hyperledger Fabric v${FABRIC_TAG} into ${FABRIC_SRC_DIR}..."
  git clone --depth 1 --branch "v${FABRIC_TAG}" git@github.com:hyperledger/fabric.git "${FABRIC_SRC_DIR}"
  echo "Building Fabric native binaries (v${FABRIC_TAG})..."
  (cd "${FABRIC_SRC_DIR}" && make native)
  mkdir -p "${HOME}/.local/bin"
  find "${FABRIC_SRC_DIR}/build/bin" -maxdepth 1 -type f -exec cp -f {} "${HOME}/.local/bin/" \;
  echo "Fabric binaries installed to ~/.local/bin — ensure it is on your PATH."
  mkdir -p "${SCRIPT_DIR}/../../vars"
  cp "${FABRIC_SRC_DIR}/sampleconfig/core.yaml" ${SCRIPT_DIR}/../../vars/core.yaml
  rm -rf "${FABRIC_SRC_DIR}"
else
  echo "Fabric peer binary found on PATH, skipping build."
fi
