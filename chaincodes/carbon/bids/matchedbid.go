package bids

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	MATCHED_BID_PREFIX = "matchedBid"
	MATCHED_BID_PVT    = "matchedBidPvt"
)

// MatchedBid struct represents matched bids with private price and multiplier
type MatchedBid struct {
	BuyBid            *BuyBid            `json:"buyBid"`
	SellBid           *SellBid           `json:"sellBid"`
	Quantity          int64              `json:"quantity"`
	PrivatePrice      *PrivatePrice      `json:"privatePrice"`
	PrivateMultiplier *PrivateMultiplier `json:"privateMultiplier"`
}

func (mb *MatchedBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetStateWithCompositeKey(stub, MATCHED_BID_PREFIX, keyAttributes, mb)
	if err != nil {
		return err
	}

	if err := mb.FetchPrivatePrice(stub); err != nil {
		return err
	}

	if err := mb.FetchPrivateMultiplier(stub); err != nil {
		return err
	}

	return nil
}

func (mb *MatchedBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if mb.BuyBid == nil || mb.SellBid == nil {
		return fmt.Errorf("BuyBid or SellBid is nil")
	}

	if mb.PrivatePrice == nil {
		return fmt.Errorf("PrivatePrice is nil")
	}

	if err := mb.PrivatePrice.ToWorldState(stub, MATCHED_BID_PVT); err != nil {
		return err
	}

	if mb.PrivateMultiplier != nil {
		if err := mb.PrivateMultiplier.ToWorldState(stub); err != nil {
			return err
		}
	}

	// Temporarily unset PrivatePrice and PrivateMultiplier not to store them in the public world state
	copyMb := *mb // Create a copy of MatchedBid to avoid modifying the original
	copyMb.PrivatePrice = nil
	copyMb.PrivateMultiplier = nil

	// We do not need duplicated data in the SellBid and BuyBid
	copyMb.BuyBid.PrivatePrice = nil // Private prices are already stored in the world state
	copyMb.SellBid.PrivatePrice = nil
	copyMb.SellBid.Credit = nil // Do not store the credit in the matched bid

	// Store MatchedBid in world state without private data and without duplicated data
	err := state.PutStateWithCompositeKey(stub, MATCHED_BID_PREFIX, copyMb.GetID(), copyMb)

	return err
}

func (mb *MatchedBid) GetID() *[][]string {
	buyBidIDs := *mb.BuyBid.GetID()
	sellBidIDs := *mb.SellBid.GetID()

	buyBidFirstID := append(buyBidIDs[0], sellBidIDs[0]...)
	sellBidFirstID := append(sellBidIDs[0], buyBidIDs[0]...)
	buyerIDPrefix := buyBidIDs[BUY_BID_ID_BUYER_AS_PREFIX]
	sellerIDPrefix := sellBidIDs[SELL_BID_ID_SELLER_AS_PREFIX]

	return &[][]string{buyBidFirstID, sellBidFirstID, buyerIDPrefix, sellerIDPrefix}
}

func (mb *MatchedBid) FetchPrivatePrice(stub shim.ChaincodeStubInterface) (err error) {
	privatePrice := &PrivatePrice{}
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") == nil {
		if err := privatePrice.FromWorldState(stub, (*mb.GetID())[0], MATCHED_BID_PVT); err == nil {
			mb.PrivatePrice = privatePrice
			return nil
		}
		return fmt.Errorf("could not get private price from world state: %v", err)
	}
	return nil
}

func (mb *MatchedBid) FetchPrivateMultiplier(stub shim.ChaincodeStubInterface) (err error) {
	privateMultiplier := &PrivateMultiplier{}
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") == nil {
		if err := privateMultiplier.FromWorldState(stub, (*mb.GetID())[0]); err == nil {
			mb.PrivateMultiplier = privateMultiplier
			return nil
		}
		return fmt.Errorf("could not get private multiplier from world state: %v", err)
	}
	return nil
}

func GetCallerMatchedBids(stub shim.ChaincodeStubInterface) ([]*MatchedBid, error) {
	callerID, err := cid.GetID(stub)
	if err != nil {
		return nil, fmt.Errorf("could not get caller ID: %v", err)
	}

	matchedBids, err := state.GetStateByPartialSecondaryIndex[MatchedBid](stub, MATCHED_BID_PREFIX, []string{callerID})
	if err != nil {
		return nil, fmt.Errorf("could not get matched bids for caller %s: %v", callerID, err)
	}

	// Load private data for each matched bid
	for _, mb := range matchedBids {
		if err := mb.FetchPrivatePrice(stub); err != nil {
			return nil, fmt.Errorf("could not fetch private price for matched bid: %v",
				err)
		}

		if err := mb.FetchPrivateMultiplier(stub); err != nil {
			return nil, fmt.Errorf("could not fetch private multiplier for matched bid: %v", err)
		}
	}

	return matchedBids, nil
}
