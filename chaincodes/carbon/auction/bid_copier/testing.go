//go:build testing
// +build testing

package bidcopier

import "github.com/johannww/phd-impl/chaincodes/carbon/bids"

// CopyBids in testing returns copies of the bids preserving the quantities at
// match time to enable testing of the matching logic.
func CopyBids(
	sellBid *bids.SellBid, buyBid *bids.BuyBid) (*bids.SellBid, *bids.BuyBid) {
	return preserveQuantitiesAtMatchTimeForTesting(sellBid, buyBid)
}

func preserveQuantitiesAtMatchTimeForTesting(
	sellBid *bids.SellBid, buyBid *bids.BuyBid) (*bids.SellBid, *bids.BuyBid) {
	sellCopy := *sellBid
	buyCopy := *buyBid

	buyCopy.PrivateQuantity = &bids.PrivateQuantity{
		AskQuantity: buyCopy.PrivateQuantity.AskQuantity,
		BidID:       buyCopy.PrivateQuantity.BidID,
	}
	return &sellCopy, &buyCopy
}
