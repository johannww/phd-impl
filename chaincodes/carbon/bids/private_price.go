package bids

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_DATA_COLLECTION = "privateDataCollection"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivatePrice is an for-the-government-only price encoded as a base64 string.
type PrivatePrice struct {
	ccstate.WorldStateReconstructor
	Price float64  `json:"price"`
	BidID []string `json:"bidID"` // This could be (Sell|Buy)bid or also MatchedBid
}

func (privateprice *PrivatePrice) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}

func (privateprice *PrivatePrice) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}

func (privateprice *PrivatePrice) GetID() []string {
	panic("not implemented") // TODO: Implement
}
