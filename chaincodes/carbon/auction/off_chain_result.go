package auction

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

// TODO:
func ProcessOffChainAuctionResult(stub shim.ChaincodeStubInterface, resultBytes []byte) error {

	independentResult := &OffChainIndepAuctionResult{}
	err := json.Unmarshal(resultBytes, &independentResult)
	if err == nil {
		return processIndependentAuctionResult(stub, independentResult)
	}
	coupledResult := &OffChainCoupledAuctionResult{}
	err = json.Unmarshal(resultBytes, &coupledResult)
	if err != nil {
		return fmt.Errorf("could not unmarshal auction result into either independent or coupled result: %v", err)
	}
	return processCoupledAuctionResult(stub, coupledResult)
}

func processIndependentAuctionResult(stub shim.ChaincodeStubInterface, result *OffChainIndepAuctionResult) error {
	return nil
}

func processCoupledAuctionResult(stub shim.ChaincodeStubInterface, result *OffChainCoupledAuctionResult) error {
	return nil
}
