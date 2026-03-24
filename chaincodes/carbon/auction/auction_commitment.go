package auction

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const (
	AUCTION_COMMITMENT_PREFIX = "auctionCommitment"
)

type AuctionCommitment struct {
	EndTimestamp string `json:"endTimestamp"` // RFC 3339 UTC timestamp of the auction end
	Hash         []byte `json:"commitment"`   // sha256 hash of the auction data
}

var _ state.WorldStateManager = (*AuctionCommitment)(nil)

func (a *AuctionCommitment) ToProto() proto.Message {
	return &pb.AuctionCommitment{EndTimestamp: a.EndTimestamp, Hash: a.Hash}
}

func (a *AuctionCommitment) FromProto(m proto.Message) error {
	pa, ok := m.(*pb.AuctionCommitment)
	if !ok {
		return fmt.Errorf("unexpected proto message type for AuctionCommitment")
	}
	a.EndTimestamp = pa.EndTimestamp
	a.Hash = pa.Hash
	return nil
}

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
