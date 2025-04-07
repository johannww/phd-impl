package carbon

import (
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
