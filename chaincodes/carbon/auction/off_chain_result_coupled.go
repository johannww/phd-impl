package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
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

	// Aggregate wallet adjustments to minimize world state updates
	walletAdjustments := make(map[string]int64)

	for _, mergedMb := range mergedMbs {
		err = mergedMb.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store merged matched bid: %v", err)
		}

		if err := transferCreditToBuyer(stub, mergedMb); err != nil {
			return err
		}

		if err := calculatePaymentsAndRefunds(mergedMb, &walletAdjustments); err != nil {
			return err
		}
	}

	// Update all adjusted wallets
	for ownerID, amount := range walletAdjustments {
		wallet := &payment.VirtualTokenWallet{OwnerID: ownerID}
		err := wallet.FromWorldState(stub, []string{ownerID})
		if err != nil {
			// If it doesn't exist, we assume the balance was 0
			wallet.Quantity = amount
		} else {
			wallet.Quantity += amount
		}

		if err := wallet.ToWorldState(stub); err != nil {
			return fmt.Errorf("failed to update wallet for owner %s: %v", ownerID, err)
		}
	}

	return nil
}

// transferCreditToBuyer transfers ownership of credits from the seller to the buyer
// and accumulates payment/refund adjustments in the provided walletAdjustments map.
// The function mutates the provided map in-place.
func transferCreditToBuyer(
	stub shim.ChaincodeStubInterface,
	mergedMb *bids.MatchedBid,
) error {
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
	err := existingBuyerCredit.FromWorldState(stub, (*buyerCredit.GetID())[0])
	if err == nil {
		buyerCredit.Quantity += existingBuyerCredit.Quantity
	}

	err = buyerCredit.ToWorldState(stub)
	if err != nil {
		return fmt.Errorf("could not materialize credit for buyer %s: %v", mergedMb.BuyBid.BuyerID, err)
	}
	return nil
}

func calculatePaymentsAndRefunds(mergedMb *bids.MatchedBid, walletAdjustments *map[string]int64) error {
	if walletAdjustments == nil {
		return fmt.Errorf("wallet adjustments map cannot be nil")
	}

	matchPrice := mergedMb.PrivatePrice.Price
	matchQuantity := mergedMb.Quantity

	// TODOHP: we have to be careful here. we must consider the extra credits from the multiplier.
	// 1. Seller payment: Sellers are paid the clearing price
	(*walletAdjustments)[mergedMb.SellBid.SellerID] += matchPrice * matchQuantity

	// 2. Buyer refund: Buyers committed at their limit price, refund difference
	originalPrice := mergedMb.BuyBid.PrivatePrice.Price
	refund := (originalPrice - matchPrice) * matchQuantity
	if refund > 0 {
		(*walletAdjustments)[mergedMb.BuyBid.BuyerID] += refund
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
