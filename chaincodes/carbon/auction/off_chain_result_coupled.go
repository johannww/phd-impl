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
	if err != nil {
		return fmt.Errorf("could not store coupled matched bids: %v", err)
	}

	err = storeAdjustedBids(stub, result)

	return err
}

func storeCoupledMatchedBids(stub shim.ChaincodeStubInterface, result *OffChainCoupledAuctionResult) error {
	mergedMbs, err := result.MergeIntoSingleMatchedBids()
	if err != nil {
		return fmt.Errorf("could not merge coupled auction results into single matched bids: %v", err)
	}

	for _, mergedMb := range mergedMbs {
		err = mergedMb.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store merged matched bid: %v", err)
		}
	}
	return nil
}

func storeAdjustedBids(stub shim.ChaincodeStubInterface, result *OffChainCoupledAuctionResult) error {
	mergedAdjustedSellBids, mergedAdjustedBuyBids := result.MergeIntoSingleAdjustedBids()

	for _, mergedAdjustedSellBid := range mergedAdjustedSellBids {
		err := mergedAdjustedSellBid.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store merged adjusted sell bid: %v", err)
		}
	}

	for _, mergedAdjustedBuyBid := range mergedAdjustedBuyBids {
		err := mergedAdjustedBuyBid.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store merged adjusted buy bid: %v", err)
		}
	}

	return nil
}
