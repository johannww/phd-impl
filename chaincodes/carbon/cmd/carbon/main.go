package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
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

// TODO: implement
func (c *CarbonContract) CreateSellBid(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// TODO: implement
func (c *CarbonContract) PublishData(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// TODO: implement
func (c *CarbonContract) MintCreditsForRange(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// TODO: implement
func (c *CarbonContract) BurnCredit(ctx contractapi.TransactionContextInterface) error {
	return credits.Burn(ctx.GetStub())
}

// TODO: implement
func (c *CarbonContract) LockAuctionSemaphore(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// TODO: implement
func (c *CarbonContract) UnlockAuctionSemaphore(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// TODO: implement
func (c *CarbonContract) CommitAndRetrieveDataForTEEAuction(ctx contractapi.TransactionContextInterface) error {
	return nil
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
