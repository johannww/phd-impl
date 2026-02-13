package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

func processIndependentAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainIndepAuctionResult, resultPvt *OffChainIndepAuctionResult,
) error {
	result, err := MergeIndependentPublicPrivateResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge independent public and private results: %v", err)
	}
	err = storeIndepMatchedBids(stub, result)
	return err
}

func storeIndepMatchedBids(stub shim.ChaincodeStubInterface, result *OffChainIndepAuctionResult) error {
	for i := range result.MatchedBids {
		mb := result.MatchedBids[i]
		err := mb.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store matched bid %d: %v", i, err)
		}
	}
	
	for i := range result.AdustedSellBids {
		sb := result.AdustedSellBids[i]
		err := sb.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store adjusted sell bid %d: %v", i, err)
		}
	}
	
	for i := range result.AdustedBuyBids {
		bb := result.AdustedBuyBids[i]
		err := bb.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store adjusted buy bid %d: %v", i, err)
		}
	}
	
	return nil
}
