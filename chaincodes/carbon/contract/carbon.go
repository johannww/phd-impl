package contract

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
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

func (c *CarbonContract) SetAuctionType(
	ctx contractapi.TransactionContextInterface,
	auctionType auction.AuctionType,
) error {
	return auctionType.ToWorldState(ctx.GetStub())
}

func (c *CarbonContract) PublishExpectedCCEPolicy(ctx contractapi.TransactionContextInterface, base64CcePolicy string) error {
	return tee.ExpectedCCEPolicyToWorldState(ctx.GetStub(), base64CcePolicy)
}

// PublishInitialTEEReport stores the initial TEE report containing the
// confidential container's public key for communication and verification
func (c *CarbonContract) PublishInitialTEEReport(ctx contractapi.TransactionContextInterface, reportJsonBytes []byte) error {
	return tee.InitialReportToWorldState(ctx.GetStub(), reportJsonBytes)
}

// CommitDataForTEEAuction commits the auction data to the world state
// Since it is a write operation, it should not return the serialized auction data.
// Thus, private data is not shared with the world state.
// To retrieve the auction data, use RetrieveDataForTEEAuction instead.
func (c *CarbonContract) CommitDataForTEEAuction(ctx contractapi.TransactionContextInterface, endRFC339Timestamp string) error {
	auctionID, err := auction.IncrementAuctionID(ctx.GetStub())
	if err != nil {
		return fmt.Errorf("could not increment auction ID: %v", err)
	}

	auctionData := &auction.AuctionData{}
	err = auctionData.RetrieveData(ctx.GetStub(), endRFC339Timestamp)
	if err != nil {
		return fmt.Errorf("could not retrieve auction data: %v", err)
	}

	auctionData.AuctionID = auctionID

	serializedAD, err := auctionData.ToSerializedAuctionData()
	if err != nil {
		return fmt.Errorf("could not serialize auction data: %v", err)
	}

	err = serializedAD.CommitmentToWorldState(ctx.GetStub(), endRFC339Timestamp)
	return err
}

// RetrieveDataForTEEAuction retrieves the auction data from the world state.
// WARN: CALL THIS AS READ-ONLY OPERATION not to expose private data.
func (c *CarbonContract) RetrieveDataForTEEAuction(ctx contractapi.TransactionContextInterface, endRFC339Timestamp string) (*auction.SerializedAuctionData, error) {
	auctionData := &auction.AuctionData{}
	err := auctionData.RetrieveData(ctx.GetStub(), endRFC339Timestamp)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve auction data: %v", err)
	}

	serializedAD, err := auctionData.ToSerializedAuctionData()
	if err != nil {
		return nil, fmt.Errorf("could not serialize auction data: %v", err)
	}

	// Verify the commitment to the world state
	err = serializedAD.CommitmentFromWorldState(ctx.GetStub(), endRFC339Timestamp)
	if !serializedAD.ValidateHash() {
		return nil, fmt.Errorf("auction data hash does not match the commitment in the world state")
	}

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

// SetActivePolicies sets the list of active policies in the world state
func (c *CarbonContract) SetActivePolicies(ctx contractapi.TransactionContextInterface, activePolicies []policies.Name) error {
	stub := ctx.GetStub()
	err := policies.SetActivePolicies(stub, activePolicies)
	return err
}

// AppendActivePolicy adds a new policy to the list of active policies in the world state
func (c *CarbonContract) AppendActivePolicy(ctx contractapi.TransactionContextInterface, policy policies.Name) error {
	stub := ctx.GetStub()
	err := policies.AppendActivePolicy(stub, policies.Name(policy))
	return err
}

// DeleteActivePolicy removes a policy from the list of active policies in the world state
func (c *CarbonContract) DeleteActivePolicy(ctx contractapi.TransactionContextInterface, policy policies.Name) error {
	stub := ctx.GetStub()
	err := policies.DeleteActivePolicy(stub, policies.Name(policy))
	return err
}

// GetActivePolicies retrieves the list of caller's matched bids.
// It uses cid.GetID function to determine the id
func (c *CarbonContract) GetCallerMatchedBids(ctx contractapi.TransactionContextInterface) ([]*bids.MatchedBid, error) {
	return bids.GetCallerMatchedBids(ctx.GetStub())
}
