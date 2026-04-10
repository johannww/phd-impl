package auction

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
)

func processIndependentAuctionResult(stub shim.ChaincodeStubInterface,
	resultPub *OffChainIndepAuctionResult, resultPvt *OffChainIndepAuctionResult,
) error {
	result, err := MergeIndependentPublicPrivateResults(resultPub, resultPvt)
	if err != nil {
		return fmt.Errorf("could not merge independent public and private results: %v", err)
	}

	stub.StartWriteBatch()

	if err := storeIndependentMatchedBids(stub, result); err != nil {
		return err
	}

	if err := storeAdjustedBidsIndep(stub, result); err != nil {
		return err
	}

	err = stub.FinishWriteBatch()

	return err
}

// storeIndependentMatchedBids persists matched bids and performs credit transfers
// and creates aggregated payment/refund UTXOs per seller/buyer per auction.
func storeIndependentMatchedBids(stub shim.ChaincodeStubInterface, result *OffChainIndepAuctionResult) error {
	// Aggregate payments and refunds per owner (seller/buyer)
	payments := make(map[string]int64) // sellerID -> total payment
	refunds := make(map[string]int64)  // buyerID -> total refund

	for i := range result.MatchedBids {
		mb := result.MatchedBids[i]
		if err := mb.ToWorldState(stub); err != nil {
			return fmt.Errorf("could not store matched bid %d: %v", i, err)
		}

		if err := transferCreditToBuyerCreditWallet(stub, mb); err != nil {
			return err
		}

		// Accumulate payments and refunds
		accumulatePaymentsAndRefunds(mb, payments, refunds)
	}

	// Create aggregated UTXOs for payments and refunds
	return createPaymentAndRefundUTXOs(stub, result.AuctionID, payments, refunds)
}

func transferCreditToBuyerCreditWallet(
	stub shim.ChaincodeStubInterface,
	mergedMb *bids.MatchedBid,
) error {
	buyerCreditWallet := &payment.VirtualTokenWallet{OwnerID: mergedMb.BuyBid.BuyerID}
	if err := buyerCreditWallet.FromWorldState(stub, (*buyerCreditWallet.GetID())[0]); err != nil {
		// If it doesn't exist, we assume the balance was 0
		buyerCreditWallet.Quantity = 0
	}

	buyerCreditWallet.Quantity += mergedMb.Quantity

	err := buyerCreditWallet.ToWorldState(stub)
	if err != nil {
		return fmt.Errorf("could not materialize credit for buyer %s: %v", mergedMb.BuyBid.BuyerID, err)
	}
	return nil
}

// storeAdjustedBidsIndep persists adjusted sell and buy bids for independent auctions.
func storeAdjustedBidsIndep(stub shim.ChaincodeStubInterface, result *OffChainIndepAuctionResult) error {
	var err error

	for i := range result.AdustedSellBids {
		sb := result.AdustedSellBids[i]
		if sb.Quantity == 0 {
			err = sb.DeleteFromWorldState(stub)
		} else {
			err = sb.ToWorldState(stub)
		}

		if err != nil {
			return fmt.Errorf("could not delete or store adjusted sell bid %d: %v", i, err)
		}

	}

	for i := range result.AdustedBuyBids {
		bb := result.AdustedBuyBids[i]
		if bb.AskQuantity == 0 {
			err = bb.DeleteFromWorldState(stub)
		} else {
			err = bb.ToWorldState(stub)
		}

		if err != nil {
			return fmt.Errorf("could not delete or store adjusted buy bid %d: %v", i, err)
		}
	}

	return nil
}
