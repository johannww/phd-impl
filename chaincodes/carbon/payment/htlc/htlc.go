package htlc

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
)

// HTLC (Hashed Time Locked Contract) is a type of smart contract that allows
// for the exchange of assets between two parties in a trustless manner.
type HTLC struct {
	Hash        [32]byte `json:"hash"`
	MatchedBids []*bids.MatchedBid
}

func CreateHTLC(stub shim.ChaincodeStubInterface, hash [32]byte, matchedBids []*bids.MatchedBid) HTLC {
	// ensure creator is the seller of the matched bids
	cID, err := cid.GetID(stub)
	if err != nil {
		panic(err)
	}

	for _, matchedBid := range matchedBids {
		// TODO:
		if matchedBid.SellBid.SellerID != cID {
			panic("creator is not the seller of the matched bids")
		}
	}

	return HTLC{Hash: hash}
}
