package auction

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"slices"

	bidcopier "github.com/johannww/phd-impl/chaincodes/carbon/auction/bid_copier"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/common"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
)

// OffChainCoupledAuctionResult holds the result of an off-chain coupled auction.
// It contains the public and private part of the matched bids.
// WARN: MatchedBidsPrivate must be sent as TransientData to avoid leaking private information.
type OffChainCoupledAuctionResult struct {
	AuctionID               uint64             `json:"auctionID"`
	MatchedBidsPublic       []*bids.MatchedBid `json:"matchedBidsPublic"`
	MatchedBidsPrivate      []*bids.MatchedBid `json:"matchedBidsPrivate"`
	AdjustedSellBidsPublic  []*bids.SellBid    `json:"adjustedSellBidsPublic"`
	AdjustedSellBidsPrivate []*bids.SellBid    `json:"adjustedSellBidsPrivate"`
	AdjustedBuyBidsPublic   []*bids.BuyBid     `json:"adjustedBuyBidsPublic"`
	AdjustedBuyBidsPrivate  []*bids.BuyBid     `json:"adjustedBuyBidsPrivate"`
}

func (r *OffChainCoupledAuctionResult) MergeIntoSingleMatchedBids() ([]*bids.MatchedBid, error) {
	if len(r.MatchedBidsPublic) != len(r.MatchedBidsPrivate) {
		return nil, fmt.Errorf("mismatched lengths between public and private matched bids")
	}
	merged := make([]*bids.MatchedBid, len(r.MatchedBidsPublic))
	for i := range r.MatchedBidsPublic {
		merged[i] = &bids.MatchedBid{
			BuyBid:            r.MatchedBidsPublic[i].BuyBid.DeepCopy(),
			SellBid:           r.MatchedBidsPublic[i].SellBid.DeepCopy(),
			Quantity:          r.MatchedBidsPublic[i].Quantity,
			PrivatePrice:      r.MatchedBidsPrivate[i].PrivatePrice,
			PrivateMultiplier: r.MatchedBidsPrivate[i].PrivateMultiplier,
		}
		merged[i].BuyBid.PrivateQuantity = r.MatchedBidsPrivate[i].BuyBid.PrivateQuantity
		merged[i].BuyBid.PrivatePrice = r.MatchedBidsPrivate[i].BuyBid.PrivatePrice
		merged[i].SellBid.PrivatePrice = r.MatchedBidsPrivate[i].SellBid.PrivatePrice
	}
	return merged, nil
}

func (r *OffChainCoupledAuctionResult) MergeIntoSingleAdjustedBids() (
	mergedAdjustedSellBids []*bids.SellBid,
	mergedAdjustedBuyBids []*bids.BuyBid,
) {

	mergedAdjustedSellBids = make([]*bids.SellBid, len(r.AdjustedSellBidsPublic))
	for i := range r.AdjustedSellBidsPublic {
		mergedAdjustedSellBids[i] = &bids.SellBid{
			SellerID:     r.AdjustedSellBidsPublic[i].SellerID,
			CreditID:     r.AdjustedSellBidsPublic[i].CreditID,
			Timestamp:    r.AdjustedSellBidsPublic[i].Timestamp,
			Quantity:     r.AdjustedSellBidsPublic[i].Quantity,
			PrivatePrice: r.AdjustedSellBidsPrivate[i].PrivatePrice,
		}
	}

	mergedAdjustedBuyBids = make([]*bids.BuyBid, len(r.AdjustedBuyBidsPublic))
	for i := range r.AdjustedBuyBidsPublic {
		mergedAdjustedBuyBids[i] = &bids.BuyBid{
			BuyerID:         r.AdjustedBuyBidsPublic[i].BuyerID,
			Timestamp:       r.AdjustedBuyBidsPublic[i].Timestamp,
			PrivateQuantity: r.AdjustedBuyBidsPrivate[i].PrivateQuantity,
			PrivatePrice:    r.AdjustedBuyBidsPrivate[i].PrivatePrice,
		}
	}

	return
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

	// Track which bids have been adjusted (participated in matching)
	adjustedSellBidsMap := make(map[int]bool)
	adjustedBuyBidsMap := make(map[int]bool)

	// 3. Match bids with highest multipliers first
	for _, mult := range multArray {
		sellBid := data.SellBids[mult.SellBidIndex]
		buyBid := data.BuyBids[mult.BuyBidIndex]

		if sellBid.Quantity == 0 || buyBid.PrivateQuantity.AskQuantity == 0 {
			continue // skip if either bid is exhausted
		}

		matchPrice, matchQuantity, cleared := calculateClearingPriceAndQuantity(sellBid, buyBid, mult.Value)
		if !cleared {
			continue // skip if no clearing price is found
		}

		// In testing environments, we preserve the quantities at match time for verification.
		// In production, we zero them out to avoid leaking information.
		sellBidEditedQuantity, buyBidEditedQuantity := bidcopier.CopyBids(sellBid, buyBid)

		matchedBidPublic, matchedBidPrivate := a.mountPublicAndPrivateMatchedBid(
			sellBidEditedQuantity,
			buyBidEditedQuantity,
			matchPrice,
			matchQuantity,
			mult.Value,
		)

		sellBid.Quantity -= matchQuantity
		buyBid.PrivateQuantity.AskQuantity -= matchQuantity

		// Mark these bids as adjusted
		adjustedSellBidsMap[mult.SellBidIndex] = true
		adjustedBuyBidsMap[mult.BuyBidIndex] = true

		public.MatchedBidsPublic = append(public.MatchedBidsPublic, matchedBidPublic)
		private.MatchedBidsPrivate = append(private.MatchedBidsPrivate, matchedBidPrivate)

	}

	// 4. Shuffle matched bids to break the link between buyers and sellers. This is important because the order of matched bids implies a higher multiplier, which can lead to privacy issues.
	shuffleMatchedBids(public.MatchedBidsPublic, private.MatchedBidsPrivate)

	// 5. Collect adjusted bids
	a.collectAdjustedBids(
		public,
		private,
		data.SellBids,
		data.BuyBids,
		adjustedSellBidsMap,
		adjustedBuyBidsMap)

	return public, private, nil
}

// collectAdjustedBids collects the bids that participated in the auction.
// It returns only the bids whose indices are marked in the provided maps.
func (a *AuctionCoupledRunner) collectAdjustedBids(
	public *OffChainCoupledAuctionResult,
	private *OffChainCoupledAuctionResult,
	sellBids []*bids.SellBid,
	buyBids []*bids.BuyBid,
	adjustedSellBidsMap map[int]bool,
	adjustedBuyBidsMap map[int]bool,
) {
	public.AdjustedSellBidsPublic = make([]*bids.SellBid, 0, len(adjustedSellBidsMap))
	public.AdjustedSellBidsPublic = make([]*bids.SellBid, 0, len(adjustedBuyBidsMap))
	private.AdjustedSellBidsPrivate = make([]*bids.SellBid, 0, len(adjustedSellBidsMap))
	private.AdjustedBuyBidsPrivate = make([]*bids.BuyBid, 0, len(adjustedBuyBidsMap))

	for i, sellBid := range sellBids {
		if !adjustedSellBidsMap[i] {
			continue
		}

		sellBidPublic := &bids.SellBid{
			SellerID:  sellBid.SellerID,
			CreditID:  sellBid.CreditID,
			Timestamp: sellBid.Timestamp,
			Quantity:  sellBid.Quantity,
		}
		sellBidPrivate := &bids.SellBid{
			PrivatePrice: sellBid.PrivatePrice,
		}
		public.AdjustedSellBidsPublic = append(public.AdjustedSellBidsPublic, sellBidPublic)
		private.AdjustedSellBidsPrivate = append(private.AdjustedSellBidsPrivate, sellBidPrivate)
	}

	for i, buyBid := range buyBids {
		if !adjustedBuyBidsMap[i] {
			continue
		}

		buyBidPrivate := &bids.BuyBid{
			PrivateQuantity: buyBid.PrivateQuantity,
			PrivatePrice:    buyBid.PrivatePrice,
		}
		buyBidPublic := &bids.BuyBid{
			BuyerID:   buyBid.BuyerID,
			Timestamp: buyBid.Timestamp,
		}

		public.AdjustedBuyBidsPublic = append(public.AdjustedBuyBidsPublic, buyBidPublic)
		private.AdjustedBuyBidsPrivate = append(private.AdjustedBuyBidsPrivate, buyBidPrivate)
	}
}

// mountPublicAndPrivateMatchedBid creates the public and private parts of a matched bid.
// The public part contains only non-sensitive information, while the private part contains
// sensitive information such as private prices and quantities.
func (a *AuctionCoupledRunner) mountPublicAndPrivateMatchedBid(
	sellBidCopy *bids.SellBid,
	buyBidCopy *bids.BuyBid,
	matchPrice, matchQuantity, multValue int64,
) (pub, pvt *bids.MatchedBid) {
	pub = &bids.MatchedBid{
		BuyBid:   buyBidCopy,
		SellBid:  sellBidCopy,
		Quantity: matchQuantity,
	}
	pvt = &bids.MatchedBid{
		BuyBid: &bids.BuyBid{
			PrivateQuantity: buyBidCopy.PrivateQuantity,
			PrivatePrice:    buyBidCopy.PrivatePrice,
		},
		SellBid: &bids.SellBid{
			PrivatePrice: sellBidCopy.PrivatePrice,
		},
		PrivatePrice: &bids.PrivatePrice{
			Price: matchPrice,
			BidID: (*pub.GetID())[0],
		},
		PrivateMultiplier: &bids.PrivateMultiplier{
			MatchingID: (*pub.GetID())[0],
			Scale:      policies.MULTIPLIER_SCALE,
			Value:      multValue,
		},
	}
	// Erase private attributes from public part
	pub.BuyBid.PrivateQuantity = nil
	pub.BuyBid.PrivatePrice = nil
	pub.SellBid.PrivatePrice = nil

	return
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

	quantity := min(sellBid.Quantity, buyBid.PrivateQuantity.AskQuantity)
	if quantity < common.QUANTITY_SCALE {
		return 0, 0, false // Minimum quantity is 1 unit
	}

	maxExtraQuantity := quantity * mult / policies.MULTIPLIER_SCALE

	acquirableQuantity := quantity + maxExtraQuantity

	var toBeAcquired int64
	if buyBid.PrivateQuantity.AskQuantity >= acquirableQuantity {
		toBeAcquired = quantity
	} else {
		// toBeAcquired = buyBid.AskQuantity / (1 + mult)
		// toBeAcquired + toBeAcquired*(mult/MULTIPLIER_SCALE) = buyBid.AskQuantity
		// toBeAcquired = buyBid.AskQuantity / (1 + mult/policies.MULTIPLIER_SCALE)
		// Re-writing the denominator:
		toBeAcquired = quantity * policies.MULTIPLIER_SCALE / (policies.MULTIPLIER_SCALE + mult)
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

func NewSingleCoupledResults(
	pubResult, pvtResult *OffChainCoupledAuctionResult,
) (*OffChainCoupledAuctionResult, error) {
	if pubResult.AuctionID != pvtResult.AuctionID {
		return nil, fmt.Errorf("auction ID mismatch between public and private results")
	}

	if len(pubResult.MatchedBidsPublic) != len(pvtResult.MatchedBidsPrivate) {
		return nil, fmt.Errorf("matched bids length mismatch between public and private results")
	}

	mergedResult := &OffChainCoupledAuctionResult{
		AuctionID:               pubResult.AuctionID,
		MatchedBidsPublic:       pubResult.MatchedBidsPublic,
		MatchedBidsPrivate:      pvtResult.MatchedBidsPrivate,
		AdjustedSellBidsPublic:  pubResult.AdjustedSellBidsPublic,
		AdjustedSellBidsPrivate: pvtResult.AdjustedSellBidsPrivate,
		AdjustedBuyBidsPublic:   pubResult.AdjustedBuyBidsPublic,
		AdjustedBuyBidsPrivate:  pvtResult.AdjustedBuyBidsPrivate,
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
