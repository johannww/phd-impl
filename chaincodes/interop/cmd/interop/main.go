package main

import (
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/interop/contract"
)

func main() {
	interopCC, err := contractapi.NewChaincode(contract.NewInteropContract())
	if err != nil {
		panic(err)
	}

	if err := interopCC.Start(); err != nil {
		panic(err)
	}
}
