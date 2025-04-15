package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	cc "github.com/johannww/phd-impl/chaincodes/carbon"
)

func readTlsCert(tlsCertPath string) []byte {
	x509Cert, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic("Failed to read TLS certificate: " + err.Error())
	}
	return x509Cert
}

func main() {
	ccId := os.Getenv("CHAINCODE_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")
	tlsDisabled := os.Getenv("CHAINCODE_TLS_DISABLED")
	tlsCertPath := os.Getenv("CHAINCODE_TLS_CERT")
	tlsKeyPath := os.Getenv("CHAINCODE_TLS_KEY")
	tlsClientCACertPath := os.Getenv("CHAINCODE_CLIENT_CA_CERT")

	if addr == "" {
		if err := shim.Start(&cc.Carbon{}); err != nil {
			panic(err)
		}
		return
	}

	tlsProps := shim.TLSProperties{
		Disabled: true,
	}

	if tlsDisabled == "false" {
		if tlsCertPath == "" || tlsKeyPath == "" || tlsClientCACertPath == "" {
			panic("TLS is enabled but required paths are not set")
		}
		tlsProps.Disabled = false
		tlsProps.Key = readTlsCert(tlsKeyPath)
		tlsProps.Cert = readTlsCert(tlsCertPath)
		tlsProps.ClientCACerts = readTlsCert(tlsClientCACertPath)
		fmt.Println("mTLS is enabled")
	}

	server := shim.ChaincodeServer{
		CCID:     ccId,
		Address:  addr,
		CC:       &cc.Carbon{},
		TLSProps: tlsProps,
		KaOpts:   nil,
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}
