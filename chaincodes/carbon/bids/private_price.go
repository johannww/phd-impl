package bids

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_PRICE_PREFIX = "privatePrice"
	PRICE_SCALE      = 1000 // Scale for prices to avoid floating point precision issues
)

// PrivatePrice is an for-the-government-only price.
// It uses fixed-point arithmetic to avoid floating point issues.
// Price might be dollars, cents, millidollars, etc.
type PrivatePrice struct {
	Price int64    `json:"price"`
	BidID []string `json:"bidID"` // This could be (Sell|Buy)bid or also MatchedBid
}

var _ state.WorldStateManagerWithExtraPrefix = (*PrivatePrice)(nil)

func (privPrice *PrivatePrice) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string, extraPrefix string) error {
	priceID := append([]string{extraPrefix}, keyAttributes...)
	err := state.GetPvtDataWithCompositeKey(stub, PVT_PRICE_PREFIX, priceID, state.BIDS_PVT_DATA_COLLECTION, privPrice)
	if err != nil {
		return err
	}
	return nil
}

func (privPrice *PrivatePrice) ToWorldState(stub shim.ChaincodeStubInterface, extraPrefix string) error {
	priceID := append([]string{extraPrefix}, (*privPrice.GetID())[0]...)
	err := state.PutPvtDataWithCompositeKey(stub, PVT_PRICE_PREFIX, priceID, state.BIDS_PVT_DATA_COLLECTION, privPrice)
	if err != nil {
		return err
	}
	return nil
}

func (privPrice *PrivatePrice) DeleteFromWorldState(stub shim.ChaincodeStubInterface, extraPrefix string) error {
	priceID := append([]string{extraPrefix}, (*privPrice.GetID())[0]...)
	err := state.DeletePvtDataWithCompositeKey(stub, PVT_PRICE_PREFIX, priceID, state.BIDS_PVT_DATA_COLLECTION)
	return err
}

func (privPrice *PrivatePrice) GetID() *[][]string {
	return &[][]string{privPrice.BidID}
}

// String returns a string representation of the PrivatePrice.
// It considers the price scale to format as floating point number.
func (privPrice *PrivatePrice) String() string {
	priceFloat := float64(privPrice.Price) / float64(PRICE_SCALE)
	return fmt.Sprintf("PrivatePrice{Price: %.3f, BidID: %v}", priceFloat, privPrice.BidID)
}
