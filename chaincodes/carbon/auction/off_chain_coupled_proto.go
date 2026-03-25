package auction

import (
	"fmt"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

var _ state.ProtoConvertible = (*OffChainCoupledAuctionResult)(nil)

// ToProto implements state.ProtoConvertible.
func (o *OffChainCoupledAuctionResult) ToProto() proto.Message {
	if o == nil {
		return &pb.OffChainCoupledAuctionResult{}
	}

	pbMatchedBidsPublic := make([]*pb.MatchedBid, len(o.MatchedBidsPublic))
	for i, mb := range o.MatchedBidsPublic {
		pbMatchedBidsPublic[i] = mb.ToProto().(*pb.MatchedBid)
	}

	pbMatchedBidsPrivate := make([]*pb.MatchedBid, len(o.MatchedBidsPrivate))
	for i, mb := range o.MatchedBidsPrivate {
		pbMatchedBidsPrivate[i] = mb.ToProto().(*pb.MatchedBid)
	}

	pbAdjustedSellBidsPublic := make([]*pb.SellBid, len(o.AdjustedSellBidsPublic))
	for i, sb := range o.AdjustedSellBidsPublic {
		pbAdjustedSellBidsPublic[i] = sb.ToProto().(*pb.SellBid)
	}

	pbAdjustedSellBidsPrivate := make([]*pb.SellBid, len(o.AdjustedSellBidsPrivate))
	for i, sb := range o.AdjustedSellBidsPrivate {
		pbAdjustedSellBidsPrivate[i] = sb.ToProto().(*pb.SellBid)
	}

	pbAdjustedBuyBidsPublic := make([]*pb.BuyBid, len(o.AdjustedBuyBidsPublic))
	for i, bb := range o.AdjustedBuyBidsPublic {
		pbAdjustedBuyBidsPublic[i] = bb.ToProto().(*pb.BuyBid)
	}

	pbAdjustedBuyBidsPrivate := make([]*pb.BuyBid, len(o.AdjustedBuyBidsPrivate))
	for i, bb := range o.AdjustedBuyBidsPrivate {
		pbAdjustedBuyBidsPrivate[i] = bb.ToProto().(*pb.BuyBid)
	}

	return &pb.OffChainCoupledAuctionResult{
		AuctionID:               o.AuctionID,
		MatchedBidsPublic:       pbMatchedBidsPublic,
		MatchedBidsPrivate:      pbMatchedBidsPrivate,
		AdjustedSellBidsPublic:  pbAdjustedSellBidsPublic,
		AdjustedSellBidsPrivate: pbAdjustedSellBidsPrivate,
		AdjustedBuyBidsPublic:   pbAdjustedBuyBidsPublic,
		AdjustedBuyBidsPrivate:  pbAdjustedBuyBidsPrivate,
	}
}

// FromProto implements state.ProtoConvertible.
func (o *OffChainCoupledAuctionResult) FromProto(m proto.Message) error {
	pbResult, ok := m.(*pb.OffChainCoupledAuctionResult)
	if !ok {
		return fmt.Errorf("unexpected proto message type for OffChainCoupledAuctionResult")
	}

	o.AuctionID = pbResult.AuctionID
	if o.AuctionID == 0 {
		return fmt.Errorf("auction ID cannot be zero, this is probably an OffChainIndepAuctionResult")
	}

	o.MatchedBidsPublic = make([]*bids.MatchedBid, len(pbResult.MatchedBidsPublic))
	for i, pbMb := range pbResult.MatchedBidsPublic {
		mb := &bids.MatchedBid{}
		if err := mb.FromProto(pbMb); err != nil {
			return fmt.Errorf("could not convert MatchedBid (public) from proto: %v", err)
		}
		o.MatchedBidsPublic[i] = mb
	}

	o.MatchedBidsPrivate = make([]*bids.MatchedBid, len(pbResult.MatchedBidsPrivate))
	for i, pbMb := range pbResult.MatchedBidsPrivate {
		mb := &bids.MatchedBid{}
		if err := mb.FromProto(pbMb); err != nil {
			return fmt.Errorf("could not convert MatchedBid (private) from proto: %v", err)
		}
		o.MatchedBidsPrivate[i] = mb
	}

	o.AdjustedSellBidsPublic = make([]*bids.SellBid, len(pbResult.AdjustedSellBidsPublic))
	for i, pbSb := range pbResult.AdjustedSellBidsPublic {
		sb := &bids.SellBid{}
		if err := sb.FromProto(pbSb); err != nil {
			return fmt.Errorf("could not convert adjusted SellBid (public) from proto: %v", err)
		}
		o.AdjustedSellBidsPublic[i] = sb
	}

	o.AdjustedSellBidsPrivate = make([]*bids.SellBid, len(pbResult.AdjustedSellBidsPrivate))
	for i, pbSb := range pbResult.AdjustedSellBidsPrivate {
		sb := &bids.SellBid{}
		if err := sb.FromProto(pbSb); err != nil {
			return fmt.Errorf("could not convert adjusted SellBid (private) from proto: %v", err)
		}
		o.AdjustedSellBidsPrivate[i] = sb
	}

	o.AdjustedBuyBidsPublic = make([]*bids.BuyBid, len(pbResult.AdjustedBuyBidsPublic))
	for i, pbBb := range pbResult.AdjustedBuyBidsPublic {
		bb := &bids.BuyBid{}
		if err := bb.FromProto(pbBb); err != nil {
			return fmt.Errorf("could not convert adjusted BuyBid (public) from proto: %v", err)
		}
		o.AdjustedBuyBidsPublic[i] = bb
	}

	o.AdjustedBuyBidsPrivate = make([]*bids.BuyBid, len(pbResult.AdjustedBuyBidsPrivate))
	for i, pbBb := range pbResult.AdjustedBuyBidsPrivate {
		bb := &bids.BuyBid{}
		if err := bb.FromProto(pbBb); err != nil {
			return fmt.Errorf("could not convert adjusted BuyBid (private) from proto: %v", err)
		}
		o.AdjustedBuyBidsPrivate[i] = bb
	}

	return nil
}
