package main

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	cc "github.com/johannww/phd-impl/chaincodes/carbon"
)

func main() {
	ccId := os.Getenv("CHAINCODE_ID")
	addr := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	if addr == "" {
		if err := shim.Start(&cc.Carbon{}); err != nil {
			panic(err)
		}
		return
	}

	server := shim.ChaincodeServer{
		CCID:    ccId,
		Address: addr,
		CC:      &cc.Carbon{},
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
		KaOpts: nil,
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}
