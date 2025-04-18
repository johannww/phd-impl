package state

// import (
// 	"encoding/json"
// 	"fmt"
//
// 	"github.com/hyperledger/fabric-chaincode-go/shim"
// )

// TODO: think this in a way to avoid phantom read conflict
// func GetStateWithPrefix(stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string) ([]byte, error) {
// 	stateIterator, metadata, err := stub.GetStateByRange(
// 		objectType,
// 		keyAttributes,
// 		, bookmark)
// 	if err != nil {
// 		return fmt.Errorf("could not get state: %v", err)
// 	}
//
// 	metadata.Bookmark
//
// 	return nil, nil
// }
