package auction

import (
	"fmt"
	"slices"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
)

type OffChainIndepAuctionResult struct {
	MatchedBids []*bids.MatchedBid `json:"matchedBids"`
}

// TODO: test
func RunIndependent(data *AuctionData) (*OffChainIndepAuctionResult, error) {
	matchedBids := make([]*bids.MatchedBid, 0)

	sellBids := data.SellBids
	buyBids := data.BuyBids

	slices.SortFunc(sellBids, func(a, b *bids.SellBid) int {
		return a.Less(b)
	})
	slices.SortFunc(buyBids, func(a, b *bids.BuyBid) int {
		return a.Less(b)
	})

	i, j := 0, len(buyBids)-1
	lastMatch := [2]int{-1, -1}
	hasCuttingPrice := buyBids[j].PrivatePrice.Price < sellBids[i].PrivatePrice.Price
	for buyBids[j].PrivatePrice.Price > sellBids[i].PrivatePrice.Price {
		matchQuantity := min(sellBids[i].Quantity, buyBids[j].AskQuantity)
		sellBids[i].Quantity -= matchQuantity
		buyBids[j].AskQuantity -= matchQuantity

		matchedBid := &bids.MatchedBid{
			BuyBidID:  (*buyBids[j].GetID())[0],
			BuyBid:    buyBids[j],
			SellBidID: (*sellBids[i].GetID())[0],
			SellBid:   sellBids[i],
			Quantity:  matchQuantity,
		}

		matchedBids = append(matchedBids, matchedBid)

		lastMatch[0] = i
		lastMatch[1] = j

		if i >= len(buyBids) || j < 0 {
			break
		}
	}

	if !hasCuttingPrice {
		return nil, fmt.Errorf("no cutting price found")
	}

	cuttingPrice := buyBids[lastMatch[0]].PrivatePrice.Price + sellBids[lastMatch[1]].PrivatePrice.Price/2

	for _, matchedBid := range matchedBids {
		matchedBid.PrivatePrice.Price = cuttingPrice
		matchedBid.PrivatePrice.BidID = (*matchedBid.GetID())[0]
	}

	return &OffChainIndepAuctionResult{
		MatchedBids: matchedBids,
	}, nil
}
