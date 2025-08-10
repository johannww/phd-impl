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
	if mb.PrivateMultiplier == nil {
		return fmt.Errorf("PrivateMultiplier is nil")
	}

	if err := mb.PrivatePrice.ToWorldState(stub, MATCHED_BID_PVT); err != nil {
		return err
	}

	if err := mb.PrivateMultiplier.ToWorldState(stub); err != nil {
		return err
	}

	// Temporarily unset PrivatePrice and PrivateMultiplier not to store them in the public world state
	tempPrice := mb.PrivatePrice
	tempMultiplier := mb.PrivateMultiplier
	mb.PrivatePrice = nil
	mb.PrivateMultiplier = nil

	// We do not need duplicated data in the SellBid and BuyBid
	mb.BuyBid.PrivatePrice = nil // Private prices are already stored in the world state
	mb.SellBid.PrivatePrice = nil
	mb.SellBid.Credit = nil // Do not store the credit in the matched bid

	// Store MatchedBid in world state without private data and without duplicated data
	err := state.PutStateWithCompositeKey(stub, MATCHED_BID_PREFIX, mb.GetID(), mb)

	// Restore PrivatePrice and PrivateMultiplier
	mb.PrivatePrice = tempPrice
	mb.PrivateMultiplier = tempMultiplier

	return err
}

func (mb *MatchedBid) GetID() *[][]string {
	buyBidFirstID := append((*mb.BuyBid.GetID())[0], (*mb.SellBid.GetID())[0]...)
	sellBidFirstID := append((*mb.SellBid.GetID())[0], (*mb.BuyBid.GetID())[0]...)
	return &[][]string{buyBidFirstID, sellBidFirstID}
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
