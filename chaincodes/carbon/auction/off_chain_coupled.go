package auction

import (
	"fmt"
	"slices"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
)

// OffChainCoupledAuctionResult holds the result of an off-chain coupled auction.
// It contains the public and private part of the matched bids.
// WARN: MatchedBidsPrivate must be sent as TransientData to avoid leaking private information.
type OffChainCoupledAuctionResult struct {
	MatchedBidsPublic  []*bids.MatchedBid `json:"matchedBidsPublic"`
	MatchedBidsPrivate []*bids.MatchedBid `json:"matchedBidsPrivate"`
}

// Multiplier represents a multiplier for a pair of bids.
// We represent as Int64 to avoid floating point precision issues.
type Multiplier struct {
	SellBidIndex int
	BuyBidIndex  int
	Value        int64
}

// RunCoupled runs an auction with coupled policies.
// TODO: implement
func RunCoupled(data *AuctionData) (*OffChainCoupledAuctionResult, error) {

	result := &OffChainCoupledAuctionResult{}

	// Based on the model from the paper
	// 1. Calculate multipliers for all bid pairs
	multArray := make([]*Multiplier, len(data.SellBids)*len(data.BuyBids))
	for sellBidIndex, sellBid := range data.SellBids {
		input := &policies.PolicyInput{
			Chunk: sellBid.Credit.Chunk,
		}
		for buyBidIndex, buyBid := range data.BuyBids {
			input.Company = data.CompaniesPvt[buyBid.BuyerID]

			multiplier, err := policies.MintCoupledMult(input, data.ActivePolicies)
			if err != nil {
				return nil, fmt.Errorf("could not calculate multiplier for sell bid %s and buy bid %s: %v", sellBid.SellerID, buyBid.BuyerID, err)
			}
			multArray[sellBidIndex*len(data.BuyBids)+buyBidIndex] = &Multiplier{
				SellBidIndex: sellBidIndex,
				BuyBidIndex:  buyBidIndex,
				Value:        multiplier,
			}
		}
	}

	// 2. Sort multipliers in descending order
	slices.SortFunc(multArray, func(a, b *Multiplier) int {
		if a.Value < b.Value {
			return -1
		} else if a.Value > b.Value {
			return 1
		}
		return 0
	})

	// 3. Match bids with highest multipliers first
	for _, mult := range multArray {
		sellBid := data.SellBids[mult.SellBidIndex]
		buyBid := data.BuyBids[mult.BuyBidIndex]

		if sellBid.Quantity == 0 || buyBid.AskQuantity == 0 {
			continue // skip if either bid is exhausted
		}

		matchPrice, matchQuantity, cleared := calculateClearingPriceAndQuantity(sellBid, buyBid, mult.Value)
		if !cleared {
			continue // skip if no clearing price is found
		}

		sellBid.Quantity -= matchQuantity
		buyBid.AskQuantity -= matchQuantity

		matchedBidPublic := &bids.MatchedBid{
			BuyBid:   buyBid,
			SellBid:  sellBid,
			Quantity: matchQuantity,
		}

		matchedBidPrivate := &bids.MatchedBid{
			PrivatePrice: &bids.PrivatePrice{
				Price: matchPrice,
				BidID: (*matchedBidPublic.GetID())[0],
			},
			PrivateMultiplier: &bids.PrivateMultiplier{
				MatchingID: (*matchedBidPublic.GetID())[0],
				Scale:      policies.MULTPLIER_SCALE,
				Value:      mult.Value,
			},
		}

		data.BuyBids[mult.BuyBidIndex] = buyBid
		data.SellBids[mult.SellBidIndex] = sellBid

		result.MatchedBidsPublic = append(result.MatchedBidsPublic, matchedBidPublic)
		result.MatchedBidsPrivate = append(result.MatchedBidsPrivate, matchedBidPrivate)

	}

	return nil, fmt.Errorf("not implemented")
}

// TODOHP: continue here
// calculateClearingPriceAndQuantity calculates the clearing price and quantity for a pair of bids.
// It considers the amount of extra credits the seller can provide based on the multiplier.
// This is implemented according to the model described on our paper.
func calculateClearingPriceAndQuantity(
	sellBid *bids.SellBid,
	buyBid *bids.BuyBid,
	mult int64) (Cp int64, Cq int64, hasClearingPrice bool) {

	quantity := min(sellBid.Quantity, buyBid.AskQuantity)
	maxExtraQuantity := quantity * mult / policies.MULTPLIER_SCALE

	acquirableQuantity := quantity + maxExtraQuantity

	var toBeAcquired int64
	if buyBid.AskQuantity >= acquirableQuantity {
		toBeAcquired = quantity
	} else {
		// toBeAcquired = buyBid.AskQuantity / (1 + mult)
		// toBeAcquired + toBeAcquired*(mult/MULTPLIER_SCALE) = buyBid.AskQuantity
		// toBeAcquired = buyBid.AskQuantity / (1 + mult/policies.MULTPLIER_SCALE)
		// Re-writing the denominator:
		toBeAcquired = buyBid.AskQuantity * policies.MULTPLIER_SCALE / (policies.MULTPLIER_SCALE + mult)
	}

	nominalQuantity := toBeAcquired
	// trueQuantity := nominalQuantity + nominalQuantity*mult

	// Buyer pays both for the nominal quantity and for the seller's extra credits
	buyerIsWillingToPayTotal := buyBid.PrivatePrice.Price * (nominalQuantity + nominalQuantity*(mult/policies.MULTPLIER_SCALE)/2)

	// How much the seller receives for the nominal quantity
	buyerIsWillingToPayPerNominalQuantity := buyerIsWillingToPayTotal / nominalQuantity

	if sellBid.PrivatePrice.Price > buyerIsWillingToPayPerNominalQuantity {
		return 0, 0, false
	}

	Cp = (sellBid.PrivatePrice.Price + buyerIsWillingToPayPerNominalQuantity) / 2
	Cq = nominalQuantity

	return Cp, Cq, true
}
