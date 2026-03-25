package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/state"
)

// ProcessOffChainAuctionResult processes off-chain auction results using strict unmarshaling
// to distinguish between independent and coupled result types.
func ProcessOffChainAuctionResult(stub shim.ChaincodeStubInterface, resultBytesPub, resultBytesPvt []byte) error {

	indepResultPub, indepResultPvt := &OffChainIndepAuctionResult{}, &OffChainIndepAuctionResult{}

	errPubIndep := state.UnmarshalStateAsStrict(resultBytesPub, indepResultPub)
	errPvtIndep := state.UnmarshalStateAsStrict(resultBytesPvt, indepResultPvt)

	if errPubIndep == nil && errPvtIndep == nil && len(indepResultPub.MatchedBids) > 0 {
		return processIndependentAuctionResult(stub, indepResultPub, indepResultPvt)
	}

	coupledResultPub, coupledResultPvt := &OffChainCoupledAuctionResult{}, &OffChainCoupledAuctionResult{}
	errPubCoupled := state.UnmarshalStateAsStrict(resultBytesPub, coupledResultPub)
	errPvtCoupled := state.UnmarshalStateAsStrict(resultBytesPvt, coupledResultPvt)

	if errPubCoupled == nil && errPvtCoupled == nil && len(coupledResultPub.MatchedBidsPublic) > 0 {
		return processCoupledAuctionResult(stub, coupledResultPub, coupledResultPvt)
	}

	return fmt.Errorf("could not unmarshal auction result into either independent "+
		"or coupled result: indep errors (%v, %v), coupled errors (%v, %v)",
		errPubIndep, errPvtIndep, errPubCoupled, errPvtCoupled)
}
