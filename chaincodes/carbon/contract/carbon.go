package contract

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/tee"
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
func (c *CarbonContract) CreateSellBid(ctx contractapi.TransactionContextInterface) (string, error) {
	return "Not Implemented Yet", nil
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

func (c *CarbonContract) PublishExpectedCCEPolicy(ctx contractapi.TransactionContextInterface, base64CcePolicy string) error {
	return tee.CCEPolicyToWorldState(ctx.GetStub(), base64CcePolicy)
}

// PublishInitialTEEReport stores the initial TEE report containing the
// confidential container's public key for communication and verification
func (c *CarbonContract) PublishInitialTEEReport(ctx contractapi.TransactionContextInterface, reportJsonBytes []byte) error {
	return tee.InitialReportToWorldState(ctx.GetStub(), reportJsonBytes)
}

func (c *CarbonContract) CommitAndRetrieveDataForTEEAuction(ctx contractapi.TransactionContextInterface, endRFC339Timestamp string) (*auction.SerializedAuctionData, error) {
	auctionData := &auction.AuctionData{}
	err := auctionData.RetrieveData(ctx.GetStub(), endRFC339Timestamp)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve auction data: %v", err)
	}

	serializedAD, err := auctionData.ToSerializedAuctionData()
	if err != nil {
		return nil, fmt.Errorf("could not serialize auction data: %v", err)
	}

	err = serializedAD.CommitmentToWorldState(ctx.GetStub(), endRFC339Timestamp)
	return serializedAD, err
}

func (c *CarbonContract) CheckCredAttr(ctx contractapi.TransactionContextInterface, attrName string) (string, error) {
	stub := ctx.GetStub()
	attrValue, found, err := cid.GetAttributeValue(stub, attrName)
	if err != nil {
		return "", err
	}

	if !found {
		return "", fmt.Errorf("Attribute '%s' not found", attrName)
	}

	return attrValue, nil
}
