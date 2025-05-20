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

type WorldStateManagerWithExtraPrefix interface {
	FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string, extraPrefix string) error
	ToWorldState(stub shim.ChaincodeStubInterface, extraPrefix string) error
	GetID() *[][]string
}
