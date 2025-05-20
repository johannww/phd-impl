package bids

import "github.com/hyperledger/fabric-chaincode-go/v2/shim"

const (
	MATCHED_BID_PREFIX = "matchedBid"
	MATCHED_BID_PVT    = "matchedBidPvt"
)

// TODO: Perhaps I should keep separate structs.
// One for the LevelDB and one for the application.
type MatchedBid struct {
	BuyBidID          []string           `json:"buyBidID"`
	BuyBid            *BuyBid            `json:"buyBid"`
	SellBidID         []string           `json:"sellBidID"`
	SellBid           *SellBid           `json:"sellBid"`
	Quantity          float64            `json:"quantity"`
	PrivatePrice      *PrivatePrice      `json:"privatePrice"`
	PrivateMultiplier *PrivateMultiplier `json:"privateMultiplier"`
}

func (mb *MatchedBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	panic("not implemented") // TODO: Implement
}
func (mb *MatchedBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	panic("not implemented") // TODO: Implement
}
func (mb *MatchedBid) GetID() *[][]string {
	buyBidFirstID := append((*mb.BuyBid.GetID())[0], (*mb.SellBid.GetID())[0]...)
	sellBidFirstID := append((*mb.SellBid.GetID())[0], (*mb.BuyBid.GetID())[0]...)
	return &[][]string{buyBidFirstID, sellBidFirstID}
}
