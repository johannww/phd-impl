package bids

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const (
	PVT_QUANTITY_PREFIX = "privateQuantity"
)

// PrivateQuantity prevents the general public from inferring
type PrivateQuantity struct {
	AskQuantity int64    `json:"quantity"`
	BidID       []string `json:"bidID"`
}

var _ state.WorldStateManager = (*PrivateQuantity)(nil)

func (privQuantity *PrivateQuantity) ToProto() proto.Message {
	return &pb.PrivateQuantity{
		AskQuantity: privQuantity.AskQuantity,
		BidID:       privQuantity.BidID,
	}
}

func (privQuantity *PrivateQuantity) FromProto(m proto.Message) error {
	pp, ok := m.(*pb.PrivateQuantity)
	if !ok {
		return fmt.Errorf("unexpected proto message type for PrivateQuantity")
	}
	privQuantity.AskQuantity = pp.AskQuantity
	privQuantity.BidID = pp.BidID
	return nil
}

func (privQuantity *PrivateQuantity) FromWorldState(
	stub shim.ChaincodeStubInterface,
	keyAttributes []string) error {
	err := state.GetPvtDataWithCompositeKey(stub, PVT_QUANTITY_PREFIX,
		keyAttributes, state.BIDS_PVT_DATA_COLLECTION, privQuantity)
	if err != nil {
		return err
	}
	return nil
}

func (privQuantity *PrivateQuantity) ToWorldState(stub shim.ChaincodeStubInterface) error {
	quantityFirstID := (*privQuantity.GetID())[0]
	err := state.PutPvtDataWithCompositeKey(stub, PVT_QUANTITY_PREFIX,
		quantityFirstID, state.BIDS_PVT_DATA_COLLECTION, privQuantity)
	if err != nil {
		return err
	}
	return nil
}

func (privQuantity *PrivateQuantity) DeleteFromWorldState(stub shim.ChaincodeStubInterface) error {
	quantityFirstID := (*privQuantity.GetID())[0]
	err := state.DeletePvtDataWithCompositeKey(stub, PVT_QUANTITY_PREFIX,
		quantityFirstID, state.BIDS_PVT_DATA_COLLECTION)
	return err
}

func (privQuantity *PrivateQuantity) GetID() *[][]string {
	return &[][]string{privQuantity.BidID}
}
