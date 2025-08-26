package auction

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

// TODO:
func ProcessOffChainAuctionResult(stub shim.ChaincodeStubInterface, resultBytesPub, resultBytesPvt []byte) error {

	indepResultPub, indepResultPvt := &OffChainIndepAuctionResult{}, &OffChainIndepAuctionResult{}
	err1 := json.Unmarshal(resultBytesPub, indepResultPub)
	err2 := json.Unmarshal(resultBytesPvt, indepResultPvt)
	if err1 == nil && err2 == nil {
		return processIndependentAuctionResult(stub, indepResultPub, indepResultPvt)
	}

	coupledResultPub, coupledResultPvt := &OffChainCoupledAuctionResult{}, &OffChainCoupledAuctionResult{}
	err1 = json.Unmarshal(resultBytesPub, coupledResultPub)
	err2 = json.Unmarshal(resultBytesPvt, coupledResultPvt)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("could not unmarshal auction result into either independent or coupled result: %v, %v", err1, err2)
	}
	return processCoupledAuctionResult(stub, coupledResultPub, coupledResultPvt)
}

func processIndependentAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainIndepAuctionResult, resultPvt *OffChainIndepAuctionResult,
) error {
	result, err := MergeIndependentPublicPrivateResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge independent public and private results: %v", err)
	}

	_ = result // TODO: do something with the result
	return nil
}

func processCoupledAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainCoupledAuctionResult, resultPvt *OffChainCoupledAuctionResult,
) error {
	result, err := MergeCoupledPublicPrivateResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge coupled public and private results: %v", err)
	}
	_ = result // TODO: do something with the result
	return nil
}
