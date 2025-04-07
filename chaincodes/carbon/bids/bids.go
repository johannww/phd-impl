package bids

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

func retractBid(stub shim.ChaincodeStubInterface, objectType string, bidID []string) error {
	if err := ccstate.DeleteStateWithCompositeKey(stub, objectType, bidID); err != nil {
		return fmt.Errorf("could not delete bid: %v", err)
	}
	if err := ccstate.DeletePvtDataWithCompositeKey(stub, objectType, bidID, PVT_DATA_COLLECTION); err != nil {
		return fmt.Errorf("could not delete private data: %v", err)
	}
	return nil
}
