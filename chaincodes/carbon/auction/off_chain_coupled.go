package auction

import (
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
)

// OffChainCoupledAuctionResult holds the result of an off-chain coupled auction.
// It contains the public and private part of the matched bids.
// WARN: MatchedBidsPrivate must be sent as TransientData to avoid leaking private information.
type OffChainCoupledAuctionResult struct {
	MatchedBidsPublic  []*bids.MatchedBid `json:"matchedBidsPublic"`
	MatchedBidsPrivate []*bids.MatchedBid `json:"matchedBidsPrivate"`
}

type Multiplier struct {
	SellBidIndex int
	BuyBidIndex  int
	Value        float64
}

// RunCoupled runs an auction with coupled policies.
// TODO: implement
// func RunCoupled(data *AuctionData) (*OffChainCoupledAuctionResult, error) {
// 	// Based on the model from the paper
// 	// 1. Calculate multipliers for all bid pairs
// 	multArray := make([]*Multiplier, len(data.SellBids)*len(data.BuyBids))
// 	for sellBidIndex, sellBid := range data.SellBids {
// 		input := &policies.PolicyInput{
// 			Chunk: sellBid.Credit.Chunk,
// 		}
// 		for buyBidIndex, buyBid := range data.BuyBids {
// 			input.Company = data.CompaniesPvt[buyBid.BuyerID]
//
// 			multiplier, err := policies.MintCoupledMult(input, data.ActivePolicies)
// 			if err != nil {
// 				return nil, fmt.Errorf("could not calculate multiplier for sell bid %s and buy bid %s: %v", sellBid.SellerID, buyBid.BuyerID, err)
// 			}
// 			multArray[sellBidIndex*len(data.BuyBids)+buyBidIndex] = &Multiplier{
// 				SellBidIndex: sellBidIndex,
// 				BuyBidIndex:  buyBidIndex,
// 				Value:        multiplier,
// 			}
// 		}
// 	}
//
// 	// 2. Sort multipliers in descending order
// 	slices.SortFunc(multArray, func(a, b *Multiplier) int {
// 		if a.Value < b.Value {
// 			return -1
// 		} else if a.Value > b.Value {
// 			return 1
// 		}
// 		return 0
// 	})
//
// 	// 3. Match bids with highest multipliers first
// 	for _, mult := range multArray {
// 		sellBid := data.SellBids[mult.SellBidIndex]
// 		buyBid := data.BuyBids[mult.BuyBidIndex]
//
// 		if sellBid.Quantity == 0 || buyBid.AskQuantity == 0 {
//
// 			continue // skip if either bid is exhausted
// 		}
//
// 		matchQuantity := min(sellBid.Quantity, buyBid.AskQuantity)
// 		sellBid.Quantity -= matchQuantity
// 		buyBid.AskQuantity -= matchQuantity
//
// 		// TODO: fix the price/quantity according to the multiplier
//
// 		matchedBidPublic := &bids.MatchedBid{
// 			BuyBidID:  (*buyBid.GetID())[0],
// 			BuyBid:    buyBid,
// 			SellBidID: (*sellBid.GetID())[0],
// 			SellBid:   sellBid,
// 			Quantity:  matchQuantity,
// 		}
//
// 		MatchedBidPrivate := &bids.MatchedBid{
// 			PrivatePrice: bids.PrivatePrice{
// 				Price: mult.Value, // Use multiplier as price
// 				BidID: (*buyBid.GetID())[0],
// 			},
// 			PrivateMultiplier: mult.Value,
// 		}
//
// 		data.BuyBids[mult.BuyBidIndex] = buyBid
// 		data.SellBids[mult.SellBidIndex] = sellBid
//
// 		data.MatchedBids = append(data.MatchedBids, matchedBid)
// 	}
//
// 	// 4. Calculate clearing price and quantity for each match
// 	// 5. Repeat until no more matches are possible
// 	return nil, fmt.Errorf("not implemented")
// }
