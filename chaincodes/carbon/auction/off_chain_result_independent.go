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

	paymentWalletAdjustments, err := storeIndependentMatchedBids(stub, result)
	if err != nil {
		return err
	}

	if err := updateWalletsForAdjustments(stub, paymentWalletAdjustments); err != nil {
		return err
	}

	if err := storeAdjustedBidsIndep(stub, result); err != nil {
		return err
	}

	err = stub.FinishWriteBatch()

	return err
}

// storeIndependentMatchedBids persists matched bids and performs credit transfers
// and payment/refund calculations. It returns the aggregated wallet adjustments
// map that the caller should persist to world state.
func storeIndependentMatchedBids(stub shim.ChaincodeStubInterface, result *OffChainIndepAuctionResult) (map[string]int64, error) {
	paymentWalletAdjustments := make(map[string]int64)

	for i := range result.MatchedBids {
		mb := result.MatchedBids[i]
		if err := mb.ToWorldState(stub); err != nil {
			return nil, fmt.Errorf("could not store matched bid %d: %v", i, err)
		}

		if err := transferCreditToBuyerCreditWallet(stub, mb); err != nil {
			return nil, err
		}

		if err := calculatePaymentsAndRefunds(mb, &paymentWalletAdjustments); err != nil {
			return nil, err
		}
	}

	return paymentWalletAdjustments, nil
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

// updateWalletsForAdjustments writes aggregated wallet adjustments to world state.
func updateWalletsForAdjustments(stub shim.ChaincodeStubInterface, walletAdjustments map[string]int64) error {
	for ownerID, amount := range walletAdjustments {
		wallet := &payment.VirtualTokenWallet{OwnerID: ownerID}
		err := wallet.FromWorldState(stub, []string{ownerID})
		if err != nil {
			// If it doesn't exist, assume the balance was 0 and set to adjustment amount
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
