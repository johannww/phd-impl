package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	BUY_BID_PREFIX = "buyBid"
	BUY_BID_PVT    = "buyBidPvt"
)

// BuyBid represents an ask from a buyer.
// Their ID could be either x509 or pseudonym-based
type BuyBid struct {
	BuyerID      carbon.Identity `json:"buyerID"`
	Timestamp    string          `json:"timestamp"`
	AskQuantity  float64         `json:"askQuantity"`
	PrivatePrice *PrivatePrice   `json:"privatePrice"`
}

func PublishBuyBid(stub shim.ChaincodeStubInterface, quantity float64, buyerID carbon.Identity) error {
	priceBytes, err := ccstate.GetTransientData(stub, "price")
	if err != nil {
		return err
	}

	price, err := strconv.ParseFloat(string(priceBytes), 64)
	if err != nil {
		return fmt.Errorf("could not parse price: %v", err)
	}

	bidTS, err := stub.GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("could not get transaction timestamp: %v", err)
	}

	buyBid := &BuyBid{
		BuyerID:     buyerID,
		Timestamp:   bidTS.String(),
		AskQuantity: quantity,
	}
	bidID := buyBid.GetID()

	privatePrice := &PrivatePrice{
		Price: float64(price),
		BidID: bidID,
	}
	buyBid.PrivatePrice = privatePrice

	if err := buyBid.ToWorldState(stub); err != nil {
		return err
	}

	return nil
}

func RetractBuyBid(stub shim.ChaincodeStubInterface, bidID []string) error {
	if err := retractBid(stub, BUY_BID_PREFIX, bidID); err != nil {
		return err
	}
	return nil
}

func (b *BuyBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := ccstate.GetStateWithCompositeKey(stub, BUY_BID_PREFIX, keyAttributes, b)
	if err != nil {
		return err
	}
	return nil
}

func (b *BuyBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if b.Timestamp == "" {
		return fmt.Errorf("timestamp is empty")
	}
	if b.BuyerID == nil {
		return fmt.Errorf("buyerID is nil")
	}
	if b.AskQuantity <= 0 {
		return fmt.Errorf("ask quantity is invalid")
	}
	if b.PrivatePrice != nil {
		err := b.PrivatePrice.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not put private price in world state: %v", err)
		}
		b.PrivatePrice = nil // Let's not store private data in the world state
	}

	if err := ccstate.PutStateWithCompositeKey(stub, BUY_BID_PREFIX, b.GetID(), b); err != nil {
		return fmt.Errorf("could put buybid in state: %v", err)
	}

	return nil
}

func (b *BuyBid) GetID() []string {
	return []string{b.BuyerID.String(), b.Timestamp}
}
