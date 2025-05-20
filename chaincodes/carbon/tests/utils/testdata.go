package utils_test

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
)

// TestData holds a list as an identity map
// The map key is a string and the value is generic interface{}
type TestData struct {
	Identities  *setup.MockIdentities
	Properties  []*properties.Property
	MintCredits []*credits.MintCredit
}

func (data *TestData) SaveToWorldState(stub shim.ChaincodeStubInterface) {
	saveToWorldState(stub, data.Properties)
	saveToWorldState(stub, data.MintCredits)
}

func saveToWorldState[T state.WorldStateManager](stub shim.ChaincodeStubInterface, data []T) {
	for _, item := range data {
		if err := item.ToWorldState(stub); err != nil {
			panic(err)
		}
	}
}
