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

type AuctionIndepRunner struct{}

// TODOHP: test TEE independent auction
func (a *AuctionIndepRunner) RunIndependent(data *AuctionData) (resultPub, resultPvt *OffChainIndepAuctionResult, err error) {
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
		return nil, nil, fmt.Errorf("no cutting price found")
	}

	cuttingPrice := (buyBids[lastMatch[0]].PrivatePrice.Price + sellBids[lastMatch[1]].PrivatePrice.Price) / 2

	for _, matchedBid := range matchedBids {
		matchedBid.PrivatePrice = &bids.PrivatePrice{
			Price: cuttingPrice,
			BidID: (*matchedBid.GetID())[0],
		}
	}

	result := &OffChainIndepAuctionResult{
		AuctionID:       data.AuctionID,
		MatchedBids:     matchedBids,
		AdustedSellBids: sellBids[:lastMatch[0]+1],
		AdustedBuyBids:  buyBids[:lastMatch[1]+1],
	}
	resultPub, resultPvt = splitIntoPublicAndPrivateIndependentResult(result)

	return resultPub, resultPvt, nil
}

func splitIntoPublicAndPrivateIndependentResult(
	result *OffChainIndepAuctionResult,
) (resultPub, resultPvt *OffChainIndepAuctionResult) {
	resultPub = result
	resultPvt = &OffChainIndepAuctionResult{
		AuctionID:       result.AuctionID,
		MatchedBids:     make([]*bids.MatchedBid, len(result.MatchedBids)),
		AdustedSellBids: make([]*bids.SellBid, len(result.AdustedSellBids)),
		AdustedBuyBids:  make([]*bids.BuyBid, len(result.AdustedBuyBids)),
	}

	// Split matched bids
	for i, matchedBid := range result.MatchedBids {
		resultPvt.MatchedBids[i] = &bids.MatchedBid{
			PrivatePrice:      matchedBid.PrivatePrice,
			PrivateMultiplier: matchedBid.PrivateMultiplier,
		}
		resultPub.MatchedBids[i].PrivatePrice = nil
		resultPub.MatchedBids[i].PrivateMultiplier = nil
	}

	// Split adjusted sell bids
	for i, sellBid := range result.AdustedSellBids {
		resultPvt.AdustedSellBids[i] = &bids.SellBid{
			PrivatePrice: sellBid.PrivatePrice,
		}
		resultPub.AdustedSellBids[i].PrivatePrice = nil
	}

	// Split adjusted buy bids
	for i, buyBid := range result.AdustedBuyBids {
		resultPvt.AdustedBuyBids[i] = &bids.BuyBid{
			PrivatePrice: buyBid.PrivatePrice,
		}
		resultPub.AdustedBuyBids[i].PrivatePrice = nil
	}

	return
}

func MergeIndependentPublicPrivateResults(
	pubResult, pvtResult *OffChainIndepAuctionResult,
) (*OffChainIndepAuctionResult, error) {
	if pubResult.AuctionID != pvtResult.AuctionID {
		return nil, fmt.Errorf("auction ID mismatch between public and private results")
	}
	if len(pubResult.MatchedBids) != len(pvtResult.MatchedBids) {
		return nil, fmt.Errorf("matched bids length mismatch between public and private results")
	}
	if len(pubResult.AdustedSellBids) != len(pvtResult.AdustedSellBids) {
		return nil, fmt.Errorf("adjusted sell bids length mismatch between public and private results")
	}
	if len(pubResult.AdustedBuyBids) != len(pvtResult.AdustedBuyBids) {
		return nil, fmt.Errorf("adjusted buy bids length mismatch between public and private results")
	}

	mergedResult := &OffChainIndepAuctionResult{
		AuctionID:       pubResult.AuctionID,
		MatchedBids:     make([]*bids.MatchedBid, len(pubResult.MatchedBids)),
		AdustedSellBids: make([]*bids.SellBid, len(pubResult.AdustedSellBids)),
		AdustedBuyBids:  make([]*bids.BuyBid, len(pubResult.AdustedBuyBids)),
	}

	// Merge matched bids
	for i := range pubResult.MatchedBids {
		mergedResult.MatchedBids[i] = &bids.MatchedBid{
			BuyBid:            pubResult.MatchedBids[i].BuyBid,
			SellBid:           pubResult.MatchedBids[i].SellBid,
			Quantity:          pubResult.MatchedBids[i].Quantity,
			PrivatePrice:      pvtResult.MatchedBids[i].PrivatePrice,
			PrivateMultiplier: pvtResult.MatchedBids[i].PrivateMultiplier,
		}
	}

	// Merge adjusted sell bids
	for i := range pubResult.AdustedSellBids {
		mergedResult.AdustedSellBids[i] = &bids.SellBid{
			SellerID:     pubResult.AdustedSellBids[i].SellerID,
			CreditID:     pubResult.AdustedSellBids[i].CreditID,
			Timestamp:    pubResult.AdustedSellBids[i].Timestamp,
			Quantity:     pubResult.AdustedSellBids[i].Quantity,
			PrivatePrice: pvtResult.AdustedSellBids[i].PrivatePrice,
		}
	}
	// Merge adjusted buy bids
	for i := range pubResult.AdustedBuyBids {
		mergedResult.AdustedBuyBids[i] = &bids.BuyBid{
			BuyerID:      pubResult.AdustedBuyBids[i].BuyerID,
			AskQuantity:  pubResult.AdustedBuyBids[i].AskQuantity,
			Timestamp:    pubResult.AdustedBuyBids[i].Timestamp,
			PrivatePrice: pvtResult.AdustedBuyBids[i].PrivatePrice,
		}
	}
	return mergedResult, nil
}
