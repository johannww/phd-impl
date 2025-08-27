package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

const (
	BUY_BID_PREFIX             = "buyBid"
	BUY_BID_PVT                = "buyBidPvt"
	BUY_BID_ID_BUYER_AS_PREFIX = 1
)

// BuyBid represents an ask from a buyer.
// Their ID could be either x509 or pseudonym-based
type BuyBid struct {
	BuyerID   string `json:"buyerID,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`

	// AskQuantity is used for public cases
	AskQuantity *int64 `json:"askQuantity,omitempty"`
	// PrivateQuantity is required for coupled auctions, when
	// the multiplier can be inferred from the AskQuantity, possibly
	// revealing the anonymous buyer
	PrivateQuantity *PrivateQuantity `json:"privateQuantity,omitempty"`

	PrivatePrice *PrivatePrice `json:"privatePrice,omitempty"`
}

var _ ccstate.WorldStateManager = (*BuyBid)(nil)

func (b *BuyBid) FetchPrivatePrice(stub shim.ChaincodeStubInterface) error {
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") == nil {
		privatePrice := &PrivatePrice{}
		err := privatePrice.FromWorldState(stub, (*b.GetID())[0], BUY_BID_PVT)
		if err != nil {
			return fmt.Errorf("could not get private price from world state: %v", err)
		}

		b.PrivatePrice = privatePrice
	}
	return nil
}

func (b *BuyBid) FetchPrivateQuantity(stub shim.ChaincodeStubInterface) error {
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") == nil {
		privateQuantity := &PrivateQuantity{}
		err := privateQuantity.FromWorldState(stub, (*b.GetID())[0])
		if err != nil {
			return fmt.Errorf("could not get private price from world state: %v", err)
		}

		b.PrivateQuantity = privateQuantity
	}
	return nil
}

func (b *BuyBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := ccstate.GetStateWithCompositeKey(stub, BUY_BID_PREFIX, keyAttributes, b)
	if err != nil {
		return err
	}

	err = b.FetchPrivatePrice(stub)
	if err != nil {
		return err
	}

	if b.AskQuantity != nil {
		return nil
	}

	err = b.FetchPrivateQuantity(stub)
	return err
}

// TODO: test for the bids mutex timestamp
func (b *BuyBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if b.Timestamp == "" {
		return fmt.Errorf("timestamp is empty")
	}
	if b.BuyerID == "" {
		return fmt.Errorf("buyerID is empty")
	}
	if !b.validQuantity() {
		return fmt.Errorf("ask quantity is invalid")
	}
	if b.PrivatePrice != nil {
		err := b.PrivatePrice.ToWorldState(stub, BUY_BID_PVT)
		if err != nil {
			return fmt.Errorf("could not put private price in world state: %v", err)
		}
	}

	copyB := *b              // Create a copy of BuyBid to avoid modifying the original
	copyB.PrivatePrice = nil // Temporarily unset PrivatePrice to avoid storing it in the public world state

	var err error
	if err = ccstate.PutStateWithCompositeKey(stub, BUY_BID_PREFIX, copyB.GetID(), copyB); err != nil {
		err = fmt.Errorf("could put buybid in state: %v", err)
	}

	return err
}

func (b *BuyBid) DeleteFromWorldState(stub shim.ChaincodeStubInterface) error {
	bidID := b.GetID()
	err := ccstate.DeleteStateWithCompositeKey(stub, BUY_BID_PREFIX, bidID)
	if err != nil {
		return fmt.Errorf("could not delete buy bid: %v", err)
	}

	b.PrivatePrice.BidID = (*bidID)[0]
	err = b.PrivatePrice.DeleteFromWorldState(stub, BUY_BID_PVT)

	if b.PrivateQuantity != nil {
		b.PrivateQuantity.DeleteFromWorldState(stub)
	}

	return err
}

func (b *BuyBid) GetID() *[][]string {
	return &[][]string{
		{b.Timestamp, b.BuyerID},
		{b.BuyerID, b.Timestamp},
	}
}

func (b *BuyBid) Less(b2 *BuyBid) int {
	if b.PrivatePrice.Price < b2.PrivatePrice.Price {
		return -1
	} else if b.PrivatePrice.Price > b2.PrivatePrice.Price {
		return 1
	}
	return 0
}

func (b *BuyBid) validQuantity() bool {
	return (b.AskQuantity != nil && *b.AskQuantity > 0) ||
		(b.PrivateQuantity != nil && b.PrivateQuantity.AskQuantity > 0)
}

func PublishBuyBidWithPublicQuanitity(stub shim.ChaincodeStubInterface, quantity int64) error {
	priceBytes, err := ccstate.GetTransientData(stub, "price")
	if err != nil {
		return err
	}

	price, err := strconv.ParseInt(string(priceBytes), 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse price: %v", err)
	}

	buyerID := identities.GetID(stub)

	bidTS, err := stub.GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("could not get transaction timestamp: %v", err)
	}
	bidTSStr := utils.TimestampRFC3339UtcString(bidTS)

	buyBid := &BuyBid{
		BuyerID:     buyerID,
		Timestamp:   bidTSStr,
		AskQuantity: &quantity,
	}
	bidID := *(buyBid.GetID())

	privatePrice := &PrivatePrice{
		Price: price,
		BidID: bidID[0],
	}
	buyBid.PrivatePrice = privatePrice

	if err := buyBid.ToWorldState(stub); err != nil {
		return err
	}

	return nil
}

func RetractBuyBid(stub shim.ChaincodeStubInterface, bidID []string) error {
	mockBid := &BuyBid{
		Timestamp: bidID[0],
		BuyerID:   bidID[1],
	}
	err := mockBid.DeleteFromWorldState(stub)
	return err
}
