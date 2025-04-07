package carbon

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func getTransientData(stub shim.ChaincodeStubInterface, key string) ([]byte, error) {
	transient, err := stub.GetTransient()
	if err != nil {
		fmt.Errorf("could not get transient data: %v", err)
	}
	transientField, ok := transient[key]
	if !ok {
		return nil, fmt.Errorf("could not find \"%s\" in transient data", key)
	}
	return transientField, nil
}

func putPvtDataWithCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string, collectionName string, pvtDataStruct T) error {
	pvtDataKey, err := stub.CreateCompositeKey(objectType, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not create composite key for pvt data: %v", err)
	}
	pvtDataBytes, err := json.Marshal(pvtDataStruct)
	if err != nil {
		return fmt.Errorf("could not marshal private data: %v", err)
	}
	if err := stub.PutPrivateData(collectionName, pvtDataKey, pvtDataBytes); err != nil {
		return fmt.Errorf("could not put private data: %v", err)
	}
	return nil
}

func putStateWithCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string, stateStruct T) error {
	stateKey, err := stub.CreateCompositeKey(objectType, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not create composite key for state: %v", err)
	}
	stateBytes, err := json.Marshal(stateStruct)
	if err != nil {
		return fmt.Errorf("could not marshal state: %v", err)
	}
	if err := stub.PutState(stateKey, stateBytes); err != nil {
		return fmt.Errorf("could not put state: %v", err)
	}
	return nil
}
