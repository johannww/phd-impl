package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
)

type CarbonContract struct {
	contractapi.Contract
}

// TODO: This is only a test function
func (c *CarbonContract) CreateBuyBid(ctx contractapi.TransactionContextInterface, key string, value string) error {
	stub := ctx.GetStub()
	err := bids.PublishBuyBid(stub, 2, &identities.X509Identity{CertID: "certid"})
	// _, err = state.GetStatesByRangeCompositeKey(stub, "buyBid", []string{"a"}, []string{"ac"})
	return err
}

func main() {
	carbonCC, err := contractapi.NewChaincode(new(CarbonContract))
	if err != nil {
		panic(err)
	}

	if err := carbonCC.Start(); err != nil {
		panic(err)
	}
}
