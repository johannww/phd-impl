package state

import "github.com/hyperledger/fabric-chaincode-go/v2/shim"

// WorldStateManager is an interface for reconstructing objects
// from the world state.
// This is useful for reconstructing objects from the world state,
// considering nested fields.
type WorldStateManager interface {
	FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error
	ToWorldState(stub shim.ChaincodeStubInterface) error
	GetID() *[][]string
}

// WorldStateManagerWithExtraPrefix is used by structs saved as private fields of others
// The extra prefix allows mapping between the type of the containing struct
// and the contained struct (which implements WorldStateManagerWithExtraPrefix)
// The relationship between BuyBid and PrivatePrice illustrate that use case.
type WorldStateManagerWithExtraPrefix interface {
	FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string, extraPrefix string) error
	ToWorldState(stub shim.ChaincodeStubInterface, extraPrefix string) error
	GetID() *[][]string
}
