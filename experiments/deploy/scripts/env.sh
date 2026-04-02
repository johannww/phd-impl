SCRIPT_DIR="$(cd -- "$(dirname -- "$0")" && pwd)"
echo $SCRIPT_DIR
ORG_DIR="${ORG_DIR:-${SCRIPT_DIR}/../vars/organizations}"
VALUES_FILE="${SCRIPT_DIR}/../helm/values.yaml"

FIRST_PEER_ORG=$(yq e '.network.organizations[] | select(.peers) | .name' $VALUES_FILE | head -1)
ORDERER_ORG=$(yq e '.network.organizations[] | select(.orderers) | .name' $VALUES_FILE | head -1)
CHANNEL=$(yq e '.network.channelName' $VALUES_FILE)
PEER_NODE_PORT=$(yq e ".network.organizations[] | select(.name == \"${FIRST_PEER_ORG}\") | .nodePortBase" $VALUES_FILE)
ORDERER_NODE_PORT=$(yq e ".network.organizations[] | select(.name == \"${ORDERER_ORG}\") | .nodePortBase" $VALUES_FILE)

CLUSTER_IP="$(minikube ip)"

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_ID=cli
export CORE_PEER_ADDRESS="${CLUSTER_IP}:${PEER_NODE_PORT}"
export CORE_PEER_LOCALMSPID="$FIRST_PEER_ORG"
export CORE_PEER_MSPCONFIGPATH="${ORG_DIR}/peerOrganizations/${FIRST_PEER_ORG}/users/User1@${FIRST_PEER_ORG}/msp"
export CORE_PEER_TLS_ROOTCERT_FILE="${ORG_DIR}/peerOrganizations/${FIRST_PEER_ORG}/peers/peer0.${FIRST_PEER_ORG}/tls/ca.crt"
export ORDERER_ADDRESS="${CLUSTER_IP}:${ORDERER_NODE_PORT}"
export ORDERER_TLS_CA="${ORG_DIR}/ordererOrganizations/${ORDERER_ORG}/orderers/orderer0.${ORDERER_ORG}/tls/ca.crt"

