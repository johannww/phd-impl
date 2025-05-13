package state

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const SECONDARY_INDEX_OBJ_TYPE = "secondaryIndex"

func GetTransientData(stub shim.ChaincodeStubInterface, key string) ([]byte, error) {
	transient, err := stub.GetTransient()
	if err != nil {
		return nil, fmt.Errorf("could not get transient data: %v", err)
	}
	transientField, ok := transient[key]
	if !ok {
		return nil, fmt.Errorf("could not find \"%s\" in transient data", key)
	}
	return transientField, nil
}

func PutPvtDataWithCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string, collectionName string, pvtDataStruct T) error {
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

func PutState[T any](stub shim.ChaincodeStubInterface, key string, stateStruct T) error {
	stateBytes, err := json.Marshal(stateStruct)
	if err != nil {
		return fmt.Errorf("could not marshal state: %v", err)
	}
	if err := stub.PutState(key, stateBytes); err != nil {
		return fmt.Errorf("could not put state: %v", err)
	}
	return nil
}

func putSecondaryIndexes(stub shim.ChaincodeStubInterface, keyAttributes *[][]string, objectType string) error {
	marshalledPrimaryKey, err := json.Marshal((*keyAttributes)[0])
	if err != nil {
		return fmt.Errorf("could not marshal primary key: %v", err)
	}

	for i := 1; i < len(*keyAttributes); i++ {
		attributes := append([]string{objectType}, (*keyAttributes)[i]...)
		stateKey, err := stub.CreateCompositeKey(SECONDARY_INDEX_OBJ_TYPE, attributes)
		if err != nil {
			return fmt.Errorf("could not create composite key for state: %v", err)
		}
		if err := stub.PutState(stateKey, marshalledPrimaryKey); err != nil {
			return fmt.Errorf("could not put secondary key: %v", err)
		}
	}
	return nil
}

func PutStateWithCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, keyAttributes *[][]string, stateStruct T) error {
	stateKey, err := stub.CreateCompositeKey(objectType, (*keyAttributes)[0])
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

	if len(*keyAttributes) == 1 {
		return nil
	}

	err = putSecondaryIndexes(stub, keyAttributes, objectType)

	return err
}

func GetPvtDataWithCompositeKey[T any](
	stub shim.ChaincodeStubInterface,
	objectType string,
	keyAttributes []string,
	collectionName string,
	pvtDataStruct T,
) error {
	stateKey, err := stub.CreateCompositeKey(objectType, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not create composite key for private data: %v", err)
	}
	stateBytes, err := stub.GetPrivateData(collectionName, stateKey)
	if err != nil {
		return fmt.Errorf("could not get private data: %v", err)
	}
	err = json.Unmarshal(stateBytes, pvtDataStruct)
	if err != nil {
		return fmt.Errorf("could not unmarshal private data: %v", err)
	}
	return nil
}

func GetState[T any](stub shim.ChaincodeStubInterface, objectType string, key string, stateStruct T) error {
	stateBytes, err := stub.GetState(key)
	if err != nil {
		return fmt.Errorf("could not get state: %v", err)
	}
	err = json.Unmarshal(stateBytes, stateStruct)
	if err != nil {
		return fmt.Errorf("could not unmarshal state: %v", err)
	}
	return nil
}

func GetStateWithCompositeKey[T any](stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string, stateStruct T) error {
	stateKey, err := stub.CreateCompositeKey(objectType, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not create composite key for state: %v", err)
	}
	stateBytes, err := stub.GetState(stateKey)
	if err != nil {
		return fmt.Errorf("could not get state: %v", err)
	}
	err = json.Unmarshal(stateBytes, stateStruct)
	if err != nil {
		return fmt.Errorf("could not unmarshal state: %v", err)
	}
	return nil
}

func DeletePvtDataWithCompositeKey(stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string, collectionName string) error {
	pvtDataKey, err := stub.CreateCompositeKey(objectType, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not create composite key for pvt data: %v", err)
	}
	if err := stub.DelPrivateData(collectionName, pvtDataKey); err != nil {
		return fmt.Errorf("could not delete private data: %v", err)
	}
	// TODO: perhaps also purge the private data
	// stub.PurgePrivateData(collectionName, pvtDataKey)
	return nil
}

func DeleteStateWithCompositeKey(stub shim.ChaincodeStubInterface, objectType string, keyAttributes []string) error {
	stateKey, err := stub.CreateCompositeKey(objectType, keyAttributes)
	if err != nil {
		return fmt.Errorf("could not create composite key for state: %v", err)
	}
	if err := stub.DelState(stateKey); err != nil {
		return fmt.Errorf("could not delete state: %v", err)
	}
	return nil
}
