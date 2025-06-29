package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const AUCTION_TYPE_KEY = "auctionType"
const AUCTION_INDEPENDENT = "independent"
const AUCTION_COUPLED = "coupled"

type AuctionType string

var _ state.WorldStateManager = (*AuctionType)(nil)

// FromWorldState implements state.WorldStateManager.
func (a *AuctionType) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetState(stub, AUCTION_TYPE_KEY, a)
}

// GetID implements state.WorldStateManager.
func (a *AuctionType) GetID() *[][]string {
	panic("unimplemented")
}

// ToWorldState implements state.WorldStateManager.
func (a *AuctionType) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if *a != AUCTION_INDEPENDENT && *a != AUCTION_COUPLED {
		return fmt.Errorf("invalid auction type: %s, must be either %s or %s", *a, AUCTION_INDEPENDENT, AUCTION_COUPLED)
	}
	return state.PutState(stub, AUCTION_TYPE_KEY, a)
}
