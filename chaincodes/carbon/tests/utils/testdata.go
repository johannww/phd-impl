package utils_test

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
)

// TestData holds a list as an identity map
// The map key is a string and the value is generic interface{}
type TestData struct {
	Identities *setup.MockIdentities
	Properties []*properties.Property
}

func (data *TestData) SaveToWorldState(stub shim.ChaincodeStubInterface) {
	for _, prop := range data.Properties {
		if err := prop.ToWorldState(stub); err != nil {
			panic(err)
		}
	}
}
