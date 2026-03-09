#!/usr/bin/env bash
# generate_sicar_cert.bash – generates a TLS cert for the sicar-mock service,
# stores it as a Kubernetes TLS Secret (mounted by the pod), and publishes the
# CA cert as a ConfigMap so chaincode pods can trust it.
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SICAR_CERT_DIR="${SCRIPT_DIR}/../../vars/sicar"
SICAR_NAMESPACE="${NAMESPACE:-fabric-experiments}"
SICAR_SERVICE_NAME="sicar-mock"
SICAR_CERT_SECRET="${SICAR_SERVICE_NAME}-tls"
SICAR_CA_CONFIGMAP="${SICAR_SERVICE_NAME}-ca"

mkdir -p "${SICAR_CERT_DIR}"

if [ ! -f "${SICAR_CERT_DIR}/server.crt" ]; then
  echo "Generating TLS cert for ${SICAR_SERVICE_NAME}..."

  cat > "${SICAR_CERT_DIR}/openssl.cnf" << EOF
[req]
req_extensions     = v3_req
distinguished_name = req_distinguished_name
prompt             = no
[req_distinguished_name]
CN = ${SICAR_SERVICE_NAME}
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SICAR_SERVICE_NAME}
DNS.2 = ${SICAR_SERVICE_NAME}.${SICAR_NAMESPACE}.svc.cluster.local
DNS.3 = ${SICAR_SERVICE_NAME}.${SICAR_NAMESPACE}.svc
DNS.4 = localhost
IP.1  = 127.0.0.1
EOF

  openssl genrsa -out "${SICAR_CERT_DIR}/server.key" 2048 2>/dev/null
  openssl req -new \
    -key "${SICAR_CERT_DIR}/server.key" \
    -config "${SICAR_CERT_DIR}/openssl.cnf" \
    -out "${SICAR_CERT_DIR}/server.csr" 2>/dev/null
  openssl x509 -req \
    -in "${SICAR_CERT_DIR}/server.csr" \
    -signkey "${SICAR_CERT_DIR}/server.key" \
    -days 3650 \
    -extensions v3_req \
    -extfile "${SICAR_CERT_DIR}/openssl.cnf" \
    -out "${SICAR_CERT_DIR}/server.crt" 2>/dev/null

  echo "  cert: ${SICAR_CERT_DIR}/server.crt"
else
  echo "SICAR TLS cert already exists, skipping generation."
fi

echo "Applying K8s Secret ${SICAR_CERT_SECRET} in namespace ${SICAR_NAMESPACE}..."
kubectl create secret tls "${SICAR_CERT_SECRET}" \
  --cert="${SICAR_CERT_DIR}/server.crt" \
  --key="${SICAR_CERT_DIR}/server.key" \
  --namespace="${SICAR_NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Applying K8s ConfigMap ${SICAR_CA_CONFIGMAP} in namespace ${SICAR_NAMESPACE}..."
kubectl create configmap "${SICAR_CA_CONFIGMAP}" \
  --from-file=ca.crt="${SICAR_CERT_DIR}/server.crt" \
  --namespace="${SICAR_NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -
