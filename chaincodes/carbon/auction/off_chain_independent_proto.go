package auction

import (
	"fmt"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

var _ state.ProtoConvertible = (*OffChainIndepAuctionResult)(nil)

// ToProto implements state.ProtoConvertible.
func (o *OffChainIndepAuctionResult) ToProto() proto.Message {
	if o == nil {
		return &pb.OffChainIndepAuctionResult{}
	}

	pbMatchedBids := make([]*pb.MatchedBid, len(o.MatchedBids))
	for i, mb := range o.MatchedBids {
		pbMatchedBids[i] = mb.ToProto().(*pb.MatchedBid)
	}

	pbAdjustedSellBids := make([]*pb.SellBid, len(o.AdustedSellBids))
	for i, sb := range o.AdustedSellBids {
		pbAdjustedSellBids[i] = sb.ToProto().(*pb.SellBid)
	}

	pbAdjustedBuyBids := make([]*pb.BuyBid, len(o.AdustedBuyBids))
	for i, bb := range o.AdustedBuyBids {
		pbAdjustedBuyBids[i] = bb.ToProto().(*pb.BuyBid)
	}

	return &pb.OffChainIndepAuctionResult{
		AuctionID:        o.AuctionID,
		MatchedBids:      pbMatchedBids,
		AdjustedSellBids: pbAdjustedSellBids,
		AdjustedBuyBids:  pbAdjustedBuyBids,
	}
}

// FromProto implements state.ProtoConvertible.
func (o *OffChainIndepAuctionResult) FromProto(m proto.Message) error {
	pbResult, ok := m.(*pb.OffChainIndepAuctionResult)
	if !ok {
		return fmt.Errorf("unexpected proto message type for OffChainIndepAuctionResult")
	}

	o.AuctionID = pbResult.AuctionID
	if o.AuctionID == 0 {
		return fmt.Errorf("auction ID cannot be zero, this is probably an OffChainCoupledAuctionResult")
	}

	o.MatchedBids = make([]*bids.MatchedBid, len(pbResult.MatchedBids))
	for i, pbMb := range pbResult.MatchedBids {
		mb := &bids.MatchedBid{}
		if err := mb.FromProto(pbMb); err != nil {
			return fmt.Errorf("could not convert MatchedBid from proto: %v", err)
		}
		o.MatchedBids[i] = mb
	}

	o.AdustedSellBids = make([]*bids.SellBid, len(pbResult.AdjustedSellBids))
	for i, pbSb := range pbResult.AdjustedSellBids {
		sb := &bids.SellBid{}
		if err := sb.FromProto(pbSb); err != nil {
			return fmt.Errorf("could not convert adjusted SellBid from proto: %v", err)
		}
		o.AdustedSellBids[i] = sb
	}

	o.AdustedBuyBids = make([]*bids.BuyBid, len(pbResult.AdjustedBuyBids))
	for i, pbBb := range pbResult.AdjustedBuyBids {
		bb := &bids.BuyBid{}
		if err := bb.FromProto(pbBb); err != nil {
			return fmt.Errorf("could not convert adjusted BuyBid from proto: %v", err)
		}
		o.AdustedBuyBids[i] = bb
	}

	return nil
}
