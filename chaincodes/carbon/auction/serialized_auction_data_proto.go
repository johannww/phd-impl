package auction

import (
	"fmt"

	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

var _ state.ProtoConvertible = (*SerializedAuctionData)(nil)

// ToProto implements state.ProtoConvertible.
func (s *SerializedAuctionData) ToProto() proto.Message {
	if s == nil {
		return &pb.SerializedAuctionData{}
	}

	return &pb.SerializedAuctionData{
		AuctionDataBytes: s.AuctionDataBytes,
		Sum:              s.Sum,
	}
}

// FromProto implements state.ProtoConvertible.
func (s *SerializedAuctionData) FromProto(m proto.Message) error {
	pbSad, ok := m.(*pb.SerializedAuctionData)
	if !ok {
		return fmt.Errorf("unexpected proto message type for SerializedAuctionData")
	}

	s.AuctionDataBytes = pbSad.AuctionDataBytes
	s.Sum = pbSad.Sum

	return nil
}
