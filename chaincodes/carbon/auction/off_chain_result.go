package auction

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

// TODOHP: implement off chain auction result processing
func ProcessOffChainAuctionResult(stub shim.ChaincodeStubInterface, resultBytesPub, resultBytesPvt []byte) error {

	indepResultPub, indepResultPvt := &OffChainIndepAuctionResult{}, &OffChainIndepAuctionResult{}

	// Ensure that OffChainCoupledAuctionResult cannot be decoded into OffChainIndepAuctionResult by disallowing unknown fields in the JSON
	customDecoder := json.NewDecoder(bytes.NewReader(resultBytesPub))
	customDecoder.DisallowUnknownFields()

	err1 := customDecoder.Decode(indepResultPub)
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
