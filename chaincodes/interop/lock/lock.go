package lock

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/interop/util"
)

func canReadCreditLockStatus(stub shim.ChaincodeStubInterface, creditID []string) error {
	if cid.AssertAttributeValue(stub, identities.InteropRelayerAttr, "true") == nil {
		return nil
	}
	if len(creditID) == 0 {
		return fmt.Errorf("creditID is required unless caller has %s attribute", identities.InteropRelayerAttr)
	}
	callerID := identities.GetID(stub)
	if callerID != creditID[0] {
		return fmt.Errorf("only the credit owner or interop relayer can check locked credits")
	}
	return nil
}

func CreditIsLocked(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	lockID string,
) (bool, error) {
	if err := canReadCreditLockStatus(stub, creditID); err != nil {
		return false, err
	}

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

func LockCredit(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	quantity int64,
	destChainID string,
) (string, error) {
	funcName := "LockCredit"
	args, err := util.MarshallInvokeArgs(funcName, creditID, quantity, destChainID)
	if err != nil {
		return "", fmt.Errorf("failed to marshall arguments for %s: %v", funcName, err)
	}

	resp := stub.InvokeChaincode(carbonCCName, args, "")
	if resp.Status != 200 {
		return "", fmt.Errorf("failed to invoke chaincode %s, function %s: %s", carbonCCName, funcName, resp.Message)
	}
	return string(resp.Payload), nil
}

func CreditIsLockedForChainID(
	stub shim.ChaincodeStubInterface,
	carbonCCName string,
	creditID []string,
	lockID string,
	destChainID string,
) (bool, error) {
	if err := canReadCreditLockStatus(stub, creditID); err != nil {
		return false, err
	}

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
	if cid.AssertAttributeValue(stub, identities.InteropRelayerAttr, "true") != nil {
		return fmt.Errorf("only the interop relayer can unlock credits: missing attribute %s", identities.InteropRelayerAttr)
	}

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
