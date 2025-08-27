package bids

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_MULTIPLIER_PREFIX = "privateMultiplier"
)

// PrivateMultiplier is an for-the-(government/audtior)-only multiplier
type PrivateMultiplier struct {
	MatchingID []string `json:"matchingID"` // This could be (Sell|Buy)bid or also MatchedBid
	Scale      int64    `json:"scale"`      // The scale factor for the multiplier
	Value      int64    `json:"multiplier"` // The multiplier value, scaled by MULTIPLIER_SCALE
}

var _ state.WorldStateManager = (*PrivateMultiplier)(nil)

func (privMultiplier *PrivateMultiplier) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetPvtDataWithCompositeKey(stub, PVT_MULTIPLIER_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, privMultiplier)
	return err
}

func (privMultiplier *PrivateMultiplier) ToWorldState(stub shim.ChaincodeStubInterface) error {
	multiplierID := (*privMultiplier.GetID())[0]
	err := state.PutPvtDataWithCompositeKey(stub, PVT_MULTIPLIER_PREFIX, multiplierID, state.BIDS_PVT_DATA_COLLECTION, privMultiplier)
	return err
}

func (privMultiplier *PrivateMultiplier) GetID() *[][]string {
	return &[][]string{privMultiplier.MatchingID}
}
