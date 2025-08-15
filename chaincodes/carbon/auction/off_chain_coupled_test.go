package auction

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/quagmt/udecimal"
)

func TestCalculateClearingPrice(t *testing.T) {
	// This function is a placeholder for the actual test implementation.
	// The test should verify that the clearing price is calculated correctly
	// based on the auction data and policies.

	floatingPrice := int64(10)
	fixedPointScale := int64(bids.PRICE_SCALE)
	fixedPointPrice := floatingPrice * fixedPointScale
	bidQuantity := int64(1000)

	sellBid := &bids.SellBid{
		// Initialize with test data
		PrivatePrice: &bids.PrivatePrice{
			Price: fixedPointPrice, // Example price
		},
		Quantity: bidQuantity,
	}
	buyBid := &bids.BuyBid{
		// Initialize with test data
		PrivatePrice: &bids.PrivatePrice{
			Price: fixedPointPrice, // Example price
		},
		AskQuantity: bidQuantity,
	}
	multiplier := int64(100) // Example multiplier value

	price, quantity, hasClearing := calculateClearingPriceAndQuantity(sellBid, buyBid, multiplier)
	t.Log("Clearing Price:", price, "Clearing Quantity:", quantity, "Has Clearing Price:", hasClearing)

	price, quantity, hasClearing = calculateClearingPriceAndQuantityUdecimal(sellBid, buyBid, multiplier)
	t.Log("Clearing Price:", price, "Clearing Quantity:", quantity, "Has Clearing Price:", hasClearing)

	sellBid.PrivatePrice.Price = floatingPrice
	buyBid.PrivatePrice.Price = floatingPrice

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

func calculateClearingPriceAndQuantityUdecimal(
	sellBid *bids.SellBid,
	buyBid *bids.BuyBid,
	mult int64) (Cp int64, Cq int64, hasClearingPrice bool) {

	quantity := udecimal.MustFromInt64(min(sellBid.Quantity, buyBid.AskQuantity), 0)
	multiplier := udecimal.MustFromInt64(mult, 0)
	multiplierScale := udecimal.MustFromInt64(policies.MULTPLIER_SCALE, 0)

	maxExtraQuantity, _ := quantity.Mul(multiplier).Div(multiplierScale)
	acquirableQuantity := quantity.Add(maxExtraQuantity)

	var toBeAcquired udecimal.Decimal
	buyAskQuantity := udecimal.MustFromInt64(buyBid.AskQuantity, 0)

	if buyAskQuantity.GreaterThanOrEqual(acquirableQuantity) {
		toBeAcquired = quantity
	} else {
		denominator := multiplierScale.Add(multiplier)
		toBeAcquired, _ = buyAskQuantity.Mul(multiplierScale).Div(denominator)
	}

	nominalQuantity := toBeAcquired

	// Buyer pays both for the nominal quantity and for the seller's extra credits
	buyerIsWillingToPayTotal := udecimal.MustFromInt64(buyBid.PrivatePrice.Price, 0).Mul(func() udecimal.Decimal {
		v := nominalQuantity.Add(func() udecimal.Decimal {
			v, _ := nominalQuantity.Mul(multiplier).Div(multiplierScale.Mul(udecimal.MustFromInt64(2, 0)))
			return v
		}())
		return v
	}())

	// How much the seller receives for the nominal quantity
	buyerIsWillingToPayPerNominalQuantity, _ := buyerIsWillingToPayTotal.Div(nominalQuantity)

	if udecimal.MustFromInt64(sellBid.PrivatePrice.Price, 0).GreaterThan(buyerIsWillingToPayPerNominalQuantity) {
		return 0, 0, false
	}

	CpDecimal, _ := udecimal.MustFromInt64(sellBid.PrivatePrice.Price, 0).Add(buyerIsWillingToPayPerNominalQuantity).Div(udecimal.MustFromInt64(2, 0))

	Cp, _ = CpDecimal.Int64()
	Cq, _ = nominalQuantity.Int64()

	return Cp, Cq, true
}
