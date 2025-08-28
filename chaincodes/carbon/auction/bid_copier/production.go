//go:build !testing
// +build !testing

package bidcopier

import "github.com/johannww/phd-impl/chaincodes/carbon/bids"

// CopyBids in production returns copies of the bids with zeroed quantities to
// avoid leaking private information.
func CopyBids(
	sellBid *bids.SellBid, buyBid *bids.BuyBid) (*bids.SellBid, *bids.BuyBid) {
	return returnZeroedQuantitiesAndCopiedPrices(sellBid, buyBid)
}

func returnZeroedQuantitiesAndCopiedPrices(
	sellBid *bids.SellBid, buyBid *bids.BuyBid) (*bids.SellBid, *bids.BuyBid) {
	sellCopy := *sellBid
	buyCopy := *buyBid

	buyCopy.PrivateQuantity = &bids.PrivateQuantity{
		AskQuantity: 0,
		BidID:       buyCopy.PrivateQuantity.BidID,
	}
	sellCopy.Quantity = 0
	return &sellCopy, &buyCopy
}
