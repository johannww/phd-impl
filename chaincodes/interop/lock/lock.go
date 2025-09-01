package lock

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

func CreditIsLocked(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string) (bool, error) {
	funcName := "CreditIsLocked"
	args := [][]byte{}
	args = append(args, []byte(funcName))
	creditIdJson, err := json.Marshal(creditID)
	if err != nil {
		return false, fmt.Errorf("failed to marshal creditID: %v", err)
	}
	args = append(args, creditIdJson)

	resp := stub.InvokeChaincode(carbonCCName, args, "")
	if resp.Status != 200 {
		return false, fmt.Errorf("failed to invoke chaincode %s, function %s: %s", carbonCCName, funcName, resp.Message)
	}
	fmt.Printf("Response from %s, function %s: %s\n", carbonCCName, funcName, string(resp.Payload))
	if resp.Payload[0] == 'f' {
		return false, nil
	}

	return true, nil
}
