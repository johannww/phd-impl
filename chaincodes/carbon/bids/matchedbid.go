package bids

// TODO: Perhaps I should keep separate structs.
// One for the LevelDB and one for the application.
type MatchedBid struct {
	BuyBidID  []string `json:"buyBidID"`
	BuyBid    *BuyBid  `json:"buyBid"`
	SellBidID []string `json:"sellBidID"`
	SellBid   *SellBid `json:"sellBid"`
	Quantity  float64  `json:"quantity"`
}
