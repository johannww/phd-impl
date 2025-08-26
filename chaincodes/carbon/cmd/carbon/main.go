package main

import (
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/carbon/contract"
)

func main() {
	carbonCC, err := contractapi.NewChaincode(contract.NewCarbonContract())
	if err != nil {
		panic(err)
	}

	if err := carbonCC.Start(); err != nil {
		panic(err)
	}
}
