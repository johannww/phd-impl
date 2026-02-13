package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

func processCoupledAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainCoupledAuctionResult, resultPvt *OffChainCoupledAuctionResult,
) error {
	result, err := NewSingleCoupledResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge coupled public and private results: %v", err)
	}
	err = storeCoupledMatchedBids(stub, result)
	return err
}

func storeCoupledMatchedBids(stub shim.ChaincodeStubInterface, result *OffChainCoupledAuctionResult) error {
	for i := range result.MatchedBidsPublic {
		mbPub := result.MatchedBidsPublic[i]
		mbPvt := result.MatchedBidsPrivate[i]
		err := mbPub.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store public part of matched bid %d: %v", i, err)
		}
		err = mbPvt.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store private part of matched bid %d: %v", i, err)
		}
	}
	return nil
}
