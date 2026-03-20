package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
)

func processCoupledAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainCoupledAuctionResult, resultPvt *OffChainCoupledAuctionResult,
) error {
	result, err := NewSingleCoupledResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge coupled public and private results: %v", err)
	}
	stub.StartWriteBatch()

	err = storeCoupledMatchedBids(stub, result)
	if err != nil {
		return fmt.Errorf("could not store coupled matched bids: %v", err)
	}

	err = storeAdjustedBids(stub, result)
	if err != nil {
		return fmt.Errorf("could not store adjusted bids: %v", err)
	}

	err = stub.FinishWriteBatch()

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

		// Transfer credit ownership to the buyer
		if mergedMb.SellBid.Credit == nil {
			err := mergedMb.SellBid.FetchCredit(stub)
			if err != nil {
				return fmt.Errorf("sell bid in merged matched bid does not have associated credit data: %v", err)
			}
		}

		buyerCredit := &credits.MintCredit{
			Credit: credits.Credit{
				OwnerID:  mergedMb.BuyBid.BuyerID,
				ChunkID:  mergedMb.SellBid.Credit.ChunkID,
				Quantity: mergedMb.Quantity,
			},
			MintMult:      mergedMb.SellBid.Credit.MintMult,
			MintTimeStamp: mergedMb.SellBid.Credit.MintTimeStamp,
		}

		// Try to load existing credit for the buyer to aggregate quantity
		existingBuyerCredit := &credits.MintCredit{}
		err = existingBuyerCredit.FromWorldState(stub, (*buyerCredit.GetID())[0])
		if err == nil {
			buyerCredit.Quantity += existingBuyerCredit.Quantity
		}

		err = buyerCredit.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not materialize credit for buyer %s: %v", mergedMb.BuyBid.BuyerID, err)
		}
	}
	return nil
}

func storeAdjustedBids(stub shim.ChaincodeStubInterface, result *OffChainCoupledAuctionResult) error {
	mergedAdjustedSellBids, mergedAdjustedBuyBids := result.MergeIntoSingleAdjustedBids()

	var err error
	for _, mergedAdjustedSellBid := range mergedAdjustedSellBids {
		if mergedAdjustedSellBid.Quantity == 0 {
			err = mergedAdjustedSellBid.DeleteFromWorldState(stub)
		} else {
			err = mergedAdjustedSellBid.ToWorldState(stub)
		}
		if err != nil {
			return fmt.Errorf("could not store merged adjusted sell bid: %v", err)
		}
	}

	for _, mergedAdjustedBuyBid := range mergedAdjustedBuyBids {
		if mergedAdjustedBuyBid.PrivateQuantity.AskQuantity == 0 {
			err = mergedAdjustedBuyBid.DeleteFromWorldState(stub)
		} else {
			err = mergedAdjustedBuyBid.ToWorldState(stub)
		}

		if err != nil {
			return fmt.Errorf("could not store merged adjusted buy bid: %v", err)
		}
	}

	return nil
}
