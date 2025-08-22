package auction

import (
	"fmt"
	"slices"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
)

// OffChainIndepAuctionResult holds the result of an independent auction run off-chain.
// It contains matched bids and adjusted sell and buy bids.
// This will be consumed by the chaincode to update the world state.
type OffChainIndepAuctionResult struct {
	AuctionID       uint64             `json:"auctionID"`
	MatchedBids     []*bids.MatchedBid `json:"matchedBids"`
	AdustedSellBids []*bids.SellBid    `json:"adjustedSellBids"`
	AdustedBuyBids  []*bids.BuyBid     `json:"adjustedBuyBids"`
}

// TODOHP: test TEE independent auction
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
	hasCuttingPrice := buyBids[j].PrivatePrice.Price >= sellBids[i].PrivatePrice.Price
	for (i < len(sellBids) && j >= 0) &&
		buyBids[j].PrivatePrice.Price >= sellBids[i].PrivatePrice.Price {

		// Skip exhausted bids
		if buyBids[j].AskQuantity == 0 {
			j--
			continue
		} else if sellBids[i].Quantity == 0 {
			i++
			continue
		}

		matchQuantity := min(sellBids[i].Quantity, buyBids[j].AskQuantity)
		sellBids[i].Quantity -= matchQuantity
		buyBids[j].AskQuantity -= matchQuantity

		matchedBid := &bids.MatchedBid{
			BuyBid:   buyBids[j],
			SellBid:  sellBids[i],
			Quantity: matchQuantity,
		}

		matchedBids = append(matchedBids, matchedBid)

		lastMatch[0] = i
		lastMatch[1] = j
	}

	if !hasCuttingPrice {
		return nil, fmt.Errorf("no cutting price found")
	}

	cuttingPrice := (buyBids[lastMatch[0]].PrivatePrice.Price + sellBids[lastMatch[1]].PrivatePrice.Price) / 2

	for _, matchedBid := range matchedBids {
		matchedBid.PrivatePrice = &bids.PrivatePrice{
			Price: cuttingPrice,
			BidID: (*matchedBid.GetID())[0],
		}
	}

	return &OffChainIndepAuctionResult{
		AuctionID:       data.AuctionID,
		MatchedBids:     matchedBids,
		AdustedSellBids: sellBids[:lastMatch[0]+1],
		AdustedBuyBids:  buyBids[:lastMatch[1]+1],
	}, nil
}
