package lock

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/interop/util"
)

func CreditIsLocked(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	lockID string,
) (bool, error) {
	funcName := "CreditIsLocked"
	args, err := util.MarshallInvokeArgs(funcName, creditID, lockID)
	if err != nil {
		return false, fmt.Errorf("failed to marshall arguments for %s: %v", funcName, err)
	}

	resp := stub.InvokeChaincode(carbonCCName, args, "")
	if resp.Status != 200 {
		return false, fmt.Errorf("failed to invoke chaincode %s, function %s: %s", carbonCCName, funcName, resp.Message)
	}
	if resp.Payload[0] == 'f' {
		return false, nil
	}

	return true, nil
}

func CreditIsLockedForChainID(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	lockID string,
	destChainID string,
) (bool, error) {
	funcName := "ChainIDCreditIsLockedFor"
	args, err := util.MarshallInvokeArgs(funcName, creditID, lockID)
	if err != nil {
		return false, fmt.Errorf("failed to marshall arguments for %s: %v", funcName, err)
	}

	resp := stub.InvokeChaincode(carbonCCName, args, "")
	if resp.Status != 200 {
		return false, fmt.Errorf("failed to invoke chaincode %s, function %s: %s", carbonCCName, funcName, resp.Message)
	}
	if !bytes.Equal(resp.Payload, []byte(destChainID)) {
		return false, nil
	}

	return true, nil
}

func UnlockCredit(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	lockID string,
) error {
	funcName := "UnlockCredit"
	args, err := util.MarshallInvokeArgs(funcName, creditID, lockID)
	if err != nil {
		return fmt.Errorf("failed to marshall arguments for %s: %v", funcName, err)
	}
	resp := stub.InvokeChaincode(carbonCCName, args, "")
	if resp.Status != 200 {
		return fmt.Errorf("failed to invoke chaincode %s, function %s: %s", carbonCCName, funcName, resp.Message)
	}
	return nil
}
