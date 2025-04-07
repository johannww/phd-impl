package bids

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_DATA_COLLECTION = "privateDataCollection"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivatePrice is an for-the-government-only price encoded as a base64 string.
type PrivatePrice struct {
	Price float64  `json:"price"`
	BidID []string `json:"bidID"` // This could be (Sell|Buy)bid or also MatchedBid
}

func retractBid(stub shim.ChaincodeStubInterface, objectType string, bidID []string) error {
	if err := ccstate.DeleteStateWithCompositeKey(stub, objectType, bidID); err != nil {
		return fmt.Errorf("could not delete bid: %v", err)
	}
	if err := ccstate.DeletePvtDataWithCompositeKey(stub, objectType, bidID, PVT_DATA_COLLECTION); err != nil {
		return fmt.Errorf("could not delete private data: %v", err)
	}
	return nil
}
