package bids

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_PRICE_PREFIX = "privatePrice"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivatePrice is an for-the-government-only price encoded as a base64 string.
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
