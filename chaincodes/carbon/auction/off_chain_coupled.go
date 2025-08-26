package auction

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"slices"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/common"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
)

// OffChainCoupledAuctionResult holds the result of an off-chain coupled auction.
// It contains the public and private part of the matched bids.
// WARN: MatchedBidsPrivate must be sent as TransientData to avoid leaking private information.
type OffChainCoupledAuctionResult struct {
	AuctionID          uint64             `json:"auctionID"`
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

type AuctionCoupledRunner struct{}

// RunCoupled runs an auction with coupled policies.
// TODO: implement
func (a *AuctionCoupledRunner) RunCoupled(data *AuctionData, pApplier policies.PolicyApplier) (public, private *OffChainCoupledAuctionResult, err error) {

	public = &OffChainCoupledAuctionResult{
		AuctionID: data.AuctionID,
	}
	private = &OffChainCoupledAuctionResult{
		AuctionID: data.AuctionID,
	}

	// data.ActivePolicies
	if len(data.ActivePolicies) == 0 {
		return nil, nil, fmt.Errorf("no active policies found for auction %d", data.AuctionID)
	}

	// Based on the model from the paper
	// 1. Calculate multipliers for all bid pairs
	multArray := make([]*Multiplier, len(data.SellBids)*len(data.BuyBids))
	for sellBidIndex, sellBid := range data.SellBids {
		input := &policies.PolicyInput{
			Chunk: sellBid.Credit.Chunk,
		}
		for buyBidIndex, buyBid := range data.BuyBids {
			input.Company = data.CompaniesPvt[buyBid.BuyerID]

			multiplier, err := pApplier.MintCoupledMult(input, data.ActivePolicies)
			if err != nil {
				return nil, nil, fmt.Errorf("could not calculate multiplier for sell bid %s and buy bid %s: %v", sellBid.SellerID, buyBid.BuyerID, err)
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

		// Copy sellbid and buybid to store how much was available when
		// the match was made
		sellBidPreservedQuantity := *sellBid
		buyBidPreservedQuantity := *buyBid

		sellBid.Quantity -= matchQuantity
		buyBid.AskQuantity -= matchQuantity

		matchedBidPublic := &bids.MatchedBid{
			BuyBid:   &buyBidPreservedQuantity,
			SellBid:  &sellBidPreservedQuantity,
			Quantity: matchQuantity,
		}

		matchedBidPrivate := &bids.MatchedBid{
			PrivatePrice: &bids.PrivatePrice{
				Price: matchPrice,
				BidID: (*matchedBidPublic.GetID())[0],
			},
			PrivateMultiplier: &bids.PrivateMultiplier{
				MatchingID: (*matchedBidPublic.GetID())[0],
				Scale:      policies.MULTIPLIER_SCALE,
				Value:      mult.Value,
			},
		}

		data.BuyBids[mult.BuyBidIndex] = buyBid
		data.SellBids[mult.SellBidIndex] = sellBid

		public.MatchedBidsPublic = append(public.MatchedBidsPublic, matchedBidPublic)
		private.MatchedBidsPrivate = append(private.MatchedBidsPrivate, matchedBidPrivate)

	}

	shuffleMatchedBids(public.MatchedBidsPublic, private.MatchedBidsPrivate)

	return public, private, nil
}

// calculateClearingPriceAndQuantity calculates the clearing price and quantity for a pair of bids.
// It considers the amount of extra credits the seller can provide based on the multiplier.
// This is implemented according to the model described on our paper.
// TODOHP: reflect: if the bid is satisfied, its multiplier is higher than other,
// Thus, the pseudonym of the buyer is probably near the pseudonym of the seller.
// This can lead to privacy issues. How can we deal with that?
// - [X] After the matching, we can shuffle the matched bids to break the link between buyer and seller.
// - We might have to hide the quantities in the bids
func calculateClearingPriceAndQuantity(
	sellBid *bids.SellBid,
	buyBid *bids.BuyBid,
	mult int64) (Cp int64, Cq int64, hasClearingPrice bool) {

	// For now, multiplier must be positive.
	if mult <= 0 {
		return 0, 0, false
	}

	quantity := min(sellBid.Quantity, buyBid.AskQuantity)
	if quantity < common.QUANTITY_SCALE {
		return 0, 0, false // Minimum quantity is 1 unit
	}

	maxExtraQuantity := quantity * mult / policies.MULTIPLIER_SCALE

	acquirableQuantity := quantity + maxExtraQuantity

	var toBeAcquired int64
	if buyBid.AskQuantity >= acquirableQuantity {
		toBeAcquired = quantity
	} else {
		// toBeAcquired = buyBid.AskQuantity / (1 + mult)
		// toBeAcquired + toBeAcquired*(mult/MULTIPLIER_SCALE) = buyBid.AskQuantity
		// toBeAcquired = buyBid.AskQuantity / (1 + mult/policies.MULTIPLIER_SCALE)
		// Re-writing the denominator:
		toBeAcquired = buyBid.AskQuantity * policies.MULTIPLIER_SCALE / (policies.MULTIPLIER_SCALE + mult)
	}

	nominalQuantity := toBeAcquired
	// trueQuantity := nominalQuantity + nominalQuantity*mult

	// Buyer pays both for the nominal quantity and for the seller's extra credits
	// Expression adjusted for fixed point representation:
	buyerIsWillingToPayTotal := buyBid.PrivatePrice.Price * (nominalQuantity + nominalQuantity*mult/(2*policies.MULTIPLIER_SCALE))

	// How much the seller receives for the nominal quantity
	buyerIsWillingToPayPerNominalQuantity := buyerIsWillingToPayTotal / nominalQuantity

	if sellBid.PrivatePrice.Price > buyerIsWillingToPayPerNominalQuantity {
		return 0, 0, false
	}

	Cp = (sellBid.PrivatePrice.Price + buyerIsWillingToPayPerNominalQuantity) / 2
	Cq = nominalQuantity

	if Cq == 0 {
		return 0, 0, false
	}

	return Cp, Cq, true
}

func MergeCoupledPublicPrivateResults(
	pubResult, pvtResult *OffChainCoupledAuctionResult,
) (*OffChainCoupledAuctionResult, error) {
	if pubResult.AuctionID != pvtResult.AuctionID {
		return nil, fmt.Errorf("auction ID mismatch between public and private results")
	}

	if len(pubResult.MatchedBidsPublic) != len(pvtResult.MatchedBidsPrivate) {
		return nil, fmt.Errorf("matched bids length mismatch between public and private results")
	}

	mergedResult := &OffChainCoupledAuctionResult{
		AuctionID:          pubResult.AuctionID,
		MatchedBidsPublic:  pubResult.MatchedBidsPublic,
		MatchedBidsPrivate: pvtResult.MatchedBidsPrivate,
	}
	return mergedResult, nil
}

// shuffleMatchedBids shuffles the matched bids in both public and private slices
// because the order of matched bids imply in a higher multiplier. This eases
// infering the buyer from the seller.
func shuffleMatchedBids(matchedBidPub []*bids.MatchedBid, matchedBidPvt []*bids.MatchedBid) {
	secureShuffle(len(matchedBidPub), func(i, j int) {
		matchedBidPub[i], matchedBidPub[j] = matchedBidPub[j], matchedBidPub[i]
		matchedBidPvt[i], matchedBidPvt[j] = matchedBidPvt[j], matchedBidPvt[i]
	})
}

// secureShuffle implements the Fisher-Yates shuffle using crypto/rand for secure randomness.
func secureShuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		jBig, _ := crand.Int(crand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		swap(i, j)
	}
}
