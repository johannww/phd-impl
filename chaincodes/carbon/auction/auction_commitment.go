package auction

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	AUCTION_COMMITMENT_PREFIX = "auctionCommitment"
)

type AuctionCommitment struct {
	EndTimestamp string `json:"endTimestamp"` // RFC 3339 UTC timestamp of the auction end
	Hash         []byte `json:"commitment"`   // sha256 hash of the auction data
}

var _ state.WorldStateManager = (*AuctionCommitment)(nil)

// FromWorldState implements state.WorldStateManager.
func (a *AuctionCommitment) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, AUCTION_COMMITMENT_PREFIX, keyAttributes, a)
}

// ToWorldState implements state.WorldStateManager.
func (a *AuctionCommitment) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, AUCTION_COMMITMENT_PREFIX, a.GetID(), a)
}

// GetID implements state.WorldStateManager.
func (a *AuctionCommitment) GetID() *[][]string {
	return &[][]string{
		{a.EndTimestamp},
	}
}
