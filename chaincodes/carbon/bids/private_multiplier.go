package bids

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const (
	PVT_MULTIPLIER_PREFIX = "privateMultiplier"
)

// PrivateMultiplier is an for-the-(government/audtior)-only multiplier
type PrivateMultiplier struct {
	MatchingID []string `json:"matchingID"` // This could be (Sell|Buy)bid or also MatchedBid
	Scale      int64    `json:"scale"`      // The scale factor for the multiplier
	Value      int64    `json:"multiplier"` // The multiplier value, scaled by MULTIPLIER_SCALE
}

var _ state.WorldStateManager = (*PrivateMultiplier)(nil)

func (privMultiplier *PrivateMultiplier) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetPvtDataWithCompositeKey(stub, PVT_MULTIPLIER_PREFIX, keyAttributes, state.AUCTION_PVT_DATA_COLLECTION, privMultiplier)
	return err
}

func (privMultiplier *PrivateMultiplier) ToWorldState(stub shim.ChaincodeStubInterface) error {
	multiplierID := (*privMultiplier.GetID())[0]
	err := state.PutPvtDataWithCompositeKey(stub, PVT_MULTIPLIER_PREFIX, multiplierID, state.AUCTION_PVT_DATA_COLLECTION, privMultiplier)
	return err
}

func (privMultiplier *PrivateMultiplier) GetID() *[][]string {
	return &[][]string{privMultiplier.MatchingID}
}

func (privMultiplier *PrivateMultiplier) ToProto() proto.Message {
	return &pb.PrivateMultiplier{
		MatchingID: privMultiplier.MatchingID,
		Scale:      privMultiplier.Scale,
		Value:      privMultiplier.Value,
	}
}

func (privMultiplier *PrivateMultiplier) FromProto(m proto.Message) error {
	pm, ok := m.(*pb.PrivateMultiplier)
	if !ok {
		return fmt.Errorf("unexpected proto message type for PrivateMultiplier")
	}
	privMultiplier.MatchingID = pm.MatchingID
	privMultiplier.Scale = pm.Scale
	privMultiplier.Value = pm.Value
	return nil
}
