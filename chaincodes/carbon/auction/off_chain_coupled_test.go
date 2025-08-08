package auction

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
)

func TestCalculateClearingPrice(t *testing.T) {
	// This function is a placeholder for the actual test implementation.
	// The test should verify that the clearing price is calculated correctly
	// based on the auction data and policies.

	sellBid := &bids.SellBid{
		// Initialize with test data
		PrivatePrice: &bids.PrivatePrice{
			Price: 10, // Example price
		},
		Quantity: 100, // Example quantity
	}
	buyBid := &bids.BuyBid{
		// Initialize with test data
		PrivatePrice: &bids.PrivatePrice{
			Price: 10, // Example price
		},
		AskQuantity: 100,
	}
	multiplier := int64(100) // Example multiplier value

	price, quantity, hasClearing := calculateClearingPriceAndQuantity(sellBid, buyBid, multiplier)
	t.Log("Clearing Price:", price, "Clearing Quantity:", quantity, "Has Clearing Price:", hasClearing)
	priceFloat, quantityFloat, hasClearing := calculateClearingPriceAndQuantityFloat(sellBid, buyBid, float64(multiplier)/float64(policies.MULTPLIER_SCALE))
	t.Log("Clearing Price:", priceFloat, "Clearing Quantity:", quantityFloat, "Has Clearing Price:", hasClearing)
	t.Fatal("TestCalculateClearingPrice not implemented yet")
}

// calculateClearingPriceAndQuantityFloat serves to compare the float64 version of the clearing price calculation
func calculateClearingPriceAndQuantityFloat(
	sellBid *bids.SellBid,
	buyBid *bids.BuyBid,
	mult float64) (Cp float64, Cq float64, hasClearingPrice bool) {

	quantity := float64(min(sellBid.Quantity, buyBid.AskQuantity))
	maxExtraQuantity := float64(quantity) * mult

	acquirableQuantity := float64(quantity) + maxExtraQuantity

	var toBeAcquired float64
	if float64(buyBid.AskQuantity) >= acquirableQuantity {
		toBeAcquired = quantity
	} else {
		// toBeAcquired = buyBid.AskQuantity / (1 + mult)
		// toBeAcquired + toBeAcquired*(mult/MULTPLIER_SCALE) = buyBid.AskQuantity
		// toBeAcquired = buyBid.AskQuantity / (1 + mult/policies.MULTPLIER_SCALE)
		// Re-writing the denominator:
		toBeAcquired = float64(buyBid.AskQuantity) / (1 + mult)
	}

	nominalQuantity := toBeAcquired
	// trueQuantity := nominalQuantity + nominalQuantity*mult

	// Buyer pays both for the nominal quantity and for the seller's extra credits
	// TODOHP: review the code below
	buyerIsWillingToPayTotal := float64(buyBid.PrivatePrice.Price) * (nominalQuantity + nominalQuantity*mult/2)

	// How much the seller receives for the nominal quantity
	buyerIsWillingToPayPerNominalQuantity := buyerIsWillingToPayTotal / nominalQuantity

	if float64(sellBid.PrivatePrice.Price) > buyerIsWillingToPayPerNominalQuantity {
		return 0, 0, false
	}

	Cp = (float64(sellBid.PrivatePrice.Price) + buyerIsWillingToPayPerNominalQuantity) / 2
	Cq = nominalQuantity

	return Cp, Cq, true
}
