package bids

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	PVT_MULTIPLIER_PREFIX = "privateMultiplier"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivateMultiplier is an for-the-government-only price encoded as a base64 string.
type PrivateMultiplier struct {
	Scale int64 `json:"scale"`      // The scale factor for the multiplier
	Value int64 `json:"multiplier"` // The multiplier value, scaled by MULTPLIER_SCALE
}

var _ state.WorldStateManagerWithExtraPrefix = (*PrivateMultiplier)(nil)

func (privPrice *PrivateMultiplier) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string, extraPrefix string) error {
	panic("not implemented")
}

func (privPrice *PrivateMultiplier) ToWorldState(stub shim.ChaincodeStubInterface, extraPrefix string) error {
	panic("not implemented")
}

func (privPrice *PrivateMultiplier) GetID() *[][]string {
	panic("not implemented")
}
