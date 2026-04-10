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

	// Aggregate payments and refunds per owner (seller/buyer)
	payments := make(map[string]int64) // sellerID -> total payment
	refunds := make(map[string]int64)  // buyerID -> total refund

	for _, mergedMb := range mergedMbs {
		err = mergedMb.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not store merged matched bid: %v", err)
		}

		if err := transferCreditToBuyer(stub, mergedMb); err != nil {
			return err
		}

		// Accumulate payments and refunds
		accumulatePaymentsAndRefunds(mergedMb, payments, refunds)
	}

	// Create aggregated UTXOs for payments and refunds
	return createPaymentAndRefundUTXOs(stub, result.AuctionID, payments, refunds)
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

// accumulatePaymentsAndRefunds calculates and accumulates payments and refunds
// for a matched bid into the provided maps.
func accumulatePaymentsAndRefunds(mb *bids.MatchedBid, payments, refunds map[string]int64) {
	matchPrice := mb.PrivatePrice.Price
	matchQuantity := mb.Quantity

	// TODOHP: we have to be careful here. we must consider the extra credits from the multiplier.
	// 1. Seller payment: Sellers are paid the clearing price
	payments[mb.SellBid.SellerID] += matchPrice * matchQuantity

	// 2. Buyer refund: Buyers committed at their limit price, refund difference
	originalPrice := mb.BuyBid.PrivatePrice.Price
	refund := (originalPrice - matchPrice) * matchQuantity
	if refund > 0 {
		refunds[mb.BuyBid.BuyerID] += refund
	}
}

// createPaymentAndRefundUTXOs creates aggregated UTXOs for all payments and refunds
// for a given auction.
func createPaymentAndRefundUTXOs(stub shim.ChaincodeStubInterface, auctionID uint64, payments, refunds map[string]int64) error {
	auctionIDStr := fmt.Sprintf("%d", auctionID)

	// Create aggregated UTXOs for payments
	for sellerID, totalPayment := range payments {
		if totalPayment > 0 {
			txID := fmt.Sprintf("%s-seller", auctionIDStr)
			err := payment.CreatePaymentUTXO(stub, sellerID, totalPayment, txID)
			if err != nil {
				return fmt.Errorf("could not create payment UTXO for seller %s: %v", sellerID, err)
			}
		}
	}

	// Create aggregated UTXOs for refunds
	for buyerID, totalRefund := range refunds {
		if totalRefund > 0 {
			txID := fmt.Sprintf("%s-buyer", auctionIDStr)
			err := payment.CreateRefundUTXO(stub, buyerID, totalRefund, txID)
			if err != nil {
				return fmt.Errorf("could not create refund UTXO for buyer %s: %v", buyerID, err)
			}
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
