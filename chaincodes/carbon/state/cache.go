package state

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// GetStateFromCache tries to fetch the state from the cache first.
// T must be the type P points to.
func GetStateFromCache[P interface {
	WorldStateManager
	*T
}, T any](stub shim.ChaincodeStubInterface, cache *map[string]P, keyAttributes []string) (P, error) {
	key, _ := stub.CreateCompositeKey("", keyAttributes)
	if val, ok := (*cache)[key]; ok {
		return val, nil
	}
	var val P = new(T)
	err := val.FromWorldState(stub, keyAttributes)
	if err != nil {
		return val, fmt.Errorf("could not get value from world state: %v", err)
	}
	(*cache)[key] = val
	return val, nil
}
