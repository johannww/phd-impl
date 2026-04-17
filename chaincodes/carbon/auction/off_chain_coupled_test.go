package auction

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/common"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/quagmt/udecimal"
	"github.com/stretchr/testify/require"
)

func TestCalculateClearingPrice(t *testing.T) {
	// This function is a placeholder for the actual test implementation.
	// The test should verify that the clearing price is calculated correctly
	// based on the auction data and policies.

	floatingPrice := int64(10)
	fixedPointScale := int64(bids.PRICE_SCALE)
	fixedPointPrice := floatingPrice * fixedPointScale
	bidQuantity := int64(1000) * common.QUANTITY_SCALE

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
		PrivateQuantity: &bids.PrivateQuantity{
			AskQuantity: bidQuantity,
		},
	}
	multiplier := int64(100) // Example multiplier value

	price, quantity, hasClearing := calculateClearingPriceAndQuantity(sellBid, buyBid, multiplier)
	t.Log("Clearing Price:", price, "Clearing Quantity:", quantity, "Has Clearing Price:", hasClearing)

	priceUdecimal, quantityUdecimal, hasClearing := calculateClearingPriceAndQuantityUdecimal(sellBid, buyBid, multiplier)
	t.Log("Clearing Price:", price, "Clearing Quantity:", quantity, "Has Clearing Price:", hasClearing)

	require.Equal(t, price, priceUdecimal)
	require.Equal(t, quantity, quantityUdecimal)

	sellBid.PrivatePrice.Price = floatingPrice
	buyBid.PrivatePrice.Price = floatingPrice

	priceFloat, quantityFloat, hasClearing := calculateClearingPriceAndQuantityFloat(sellBid, buyBid, float64(multiplier)/float64(policies.MULTIPLIER_SCALE))
	t.Log("Clearing Price:", priceFloat, "Clearing Quantity:", quantityFloat, "Has Clearing Price:", hasClearing)
}

// TestCalculateClearingPriceAtMinimumQuantity tests the truncation fix at the quantity floor
func TestCalculateClearingPriceAtMinimumQuantity(t *testing.T) {
	// Test at the minimum quantity boundary (1 credit = QUANTITY_SCALE)
	minQuantity := int64(common.QUANTITY_SCALE)
	price := int64(1000) * bids.PRICE_SCALE // $1000 per credit
	mult := int64(10)                       // 1% multiplier (small to test truncation)

	sellBid := &bids.SellBid{
		PrivatePrice: &bids.PrivatePrice{Price: price},
		Quantity:     minQuantity,
	}
	buyBid := &bids.BuyBid{
		PrivatePrice: &bids.PrivatePrice{Price: price},
		PrivateQuantity: &bids.PrivateQuantity{
			AskQuantity: minQuantity * 2, // Ensure buyer isn't the limiting factor
		},
	}

	// Run all three implementations
	priceInt, quantityInt, hasClearingInt := calculateClearingPriceAndQuantity(sellBid, buyBid, mult)
	priceUdec, quantityUdec, hasClearingUdec := calculateClearingPriceAndQuantityUdecimal(sellBid, buyBid, mult)

	// For float comparison, use the unscaled price and real multiplier
	sellBid.PrivatePrice.Price = price / bids.PRICE_SCALE
	buyBid.PrivatePrice.Price = price / bids.PRICE_SCALE
	priceFloat, quantityFloat, hasClearingFloat := calculateClearingPriceAndQuantityFloat(
		sellBid, buyBid,
		float64(mult)/float64(policies.MULTIPLIER_SCALE),
	)

	// All should find a clearing price
	require.True(t, hasClearingInt, "Integer implementation should find clearing price")
	require.True(t, hasClearingUdec, "Udecimal implementation should find clearing price")
	require.True(t, hasClearingFloat, "Float implementation should find clearing price")

	// Integer and udecimal results should match exactly (this would fail with the old truncation bug)
	require.Equal(t, priceUdec, priceInt, "Prices should match at minimum quantity")
	require.Equal(t, quantityUdec, quantityInt, "Quantities should match at minimum quantity")

	// Float should be close (within rounding error)
	// Convert integer price back to float for comparison
	priceIntAsFloat := float64(priceInt) / float64(bids.PRICE_SCALE)
	require.InDelta(t, priceFloat, priceIntAsFloat, 0.01, "Float price should be close to integer price")
	require.InDelta(t, quantityFloat, float64(quantityInt), 1.0, "Float quantity should be close to integer quantity")

	// Quantity should be exactly at the floor
	require.Equal(t, minQuantity, quantityInt, "Quantity should be at minimum")

	t.Logf("At min quantity (%d), mult=%d: Cp=%d, Cq=%d (float: Cp=%.2f, Cq=%.2f)",
		minQuantity, mult, priceInt, quantityInt, priceFloat, quantityFloat)
}

// calculateClearingPriceAndQuantityFloat serves to compare the float64 version of the clearing price calculation
func calculateClearingPriceAndQuantityFloat(
	sellBid *bids.SellBid,
	buyBid *bids.BuyBid,
	mult float64) (Cp float64, Cq float64, hasClearingPrice bool) {

	quantity := float64(min(sellBid.Quantity, buyBid.PrivateQuantity.AskQuantity))
	maxExtraQuantity := float64(quantity) * mult

	acquirableQuantity := float64(quantity) + maxExtraQuantity

	var toBeAcquired float64
	if float64(buyBid.PrivateQuantity.AskQuantity) >= acquirableQuantity {
		toBeAcquired = quantity
	} else {
		// toBeAcquired = buyBid.AskQuantity / (1 + mult)
		// toBeAcquired + toBeAcquired*(mult/MULTIPLIER_SCALE) = buyBid.AskQuantity
		// toBeAcquired = buyBid.AskQuantity / (1 + mult/policies.MULTIPLIER_SCALE)
		// Re-writing the denominator:
		toBeAcquired = float64(buyBid.PrivateQuantity.AskQuantity) / (1 + mult)
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

	quantity := udecimal.MustFromInt64(min(sellBid.Quantity, buyBid.PrivateQuantity.AskQuantity), 0)
	multiplier := udecimal.MustFromInt64(mult, 0)
	multiplierScale := udecimal.MustFromInt64(policies.MULTIPLIER_SCALE, 0)

	maxExtraQuantity, _ := quantity.Mul(multiplier).Div(multiplierScale)
	acquirableQuantity := quantity.Add(maxExtraQuantity)

	var toBeAcquired udecimal.Decimal
	buyAskQuantity := udecimal.MustFromInt64(buyBid.PrivateQuantity.AskQuantity, 0)

	if buyAskQuantity.GreaterThanOrEqual(acquirableQuantity) {
		toBeAcquired = quantity
	} else {
		denominator := multiplierScale.Add(multiplier)
		toBeAcquired, _ = buyAskQuantity.Mul(multiplierScale).Div(denominator)
	}

	nominalQuantity := toBeAcquired

	// Buyer pays both for the nominal quantity and for the seller's extra credits
	// Multiply first, divide last to avoid premature truncation
	buyerIsWillingToPayTotal, _ := udecimal.MustFromInt64(buyBid.PrivatePrice.Price, 0).
		Mul(nominalQuantity).
		Mul(multiplierScale.Mul(udecimal.MustFromInt64(2, 0)).Add(multiplier)).
		Div(multiplierScale.Mul(udecimal.MustFromInt64(2, 0)))

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
