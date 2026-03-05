#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
FABRIC_TAG="${FABRIC_TAG:-3.1.4}"
IMAGE_TAG="${IMAGE_TAG:-ghcr.io/hyperledger/fabric-tools:${FABRIC_TAG}}"

DOCKER_BUILDKIT=1 docker build \
  --pull \
  --build-arg FABRIC_VER="${FABRIC_TAG}" \
  -t "${IMAGE_TAG}" \
  -f "${SCRIPT_DIR}/Dockerfile" \
  "${SCRIPT_DIR}"

echo "Built ${IMAGE_TAG}"
