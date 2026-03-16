package utils

// This are compatible with the consts defined in the fabric-contract-api
// https://github.com/hyperledger/fabric-contract-api-go/blob/b8b28e7c4a1394f5565f0d43ec3190003add14cd/contractapi/contract_chaincode.go#L47
const (
	ServerAddressVariable = "CHAINCODE_SERVER_ADDRESS"
	ChaincodeIdVariable   = "CORE_CHAINCODE_ID_NAME"
	TlsEnabledVariable    = "CORE_PEER_TLS_ENABLED"
	RootCertVariable      = "CORE_PEER_TLS_ROOTCERT_FILE"
	ClientKeyVariable     = "CORE_TLS_CLIENT_KEY_FILE"
	ClientCertVariable    = "CORE_TLS_CLIENT_CERT_FILE"
)
