package bids

const (
	PVT_DATA_COLLECTION = "privateDataCollection"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivatePrice is an for-the-government-only price encoded as a base64 string.
type PrivatePrice struct {
	Price float64  `json:"price"`
	BidID []string `json:"bidID"` // This could be (Sell|Buy)bid or also MatchedBid
}
