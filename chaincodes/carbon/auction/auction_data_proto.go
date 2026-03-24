package auction

import (
	"fmt"
	"sort"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

var _ state.ProtoConvertible = (*AuctionData)(nil)

// ToProto implements state.ProtoConvertible.
func (a *AuctionData) ToProto() proto.Message {
	if a == nil {
		return &pb.AuctionData{}
	}

	pbSellBids := make([]*pb.SellBid, len(a.SellBids))
	for i, sellBid := range a.SellBids {
		pbSellBids[i] = sellBid.ToProto().(*pb.SellBid)
	}

	pbBuyBids := make([]*pb.BuyBid, len(a.BuyBids))
	for i, buyBid := range a.BuyBids {
		pbBuyBids[i] = buyBid.ToProto().(*pb.BuyBid)
	}

	pbActivePolicies := make([]string, len(a.ActivePolicies))
	for i, policy := range a.ActivePolicies {
		pbActivePolicies[i] = string(policy)
	}

	// Convert CompaniesPvt map with sorted keys for deterministic serialization.
	// Go map iteration order is randomized, which would cause non-deterministic
	// proto marshaling. We sort the keys first to ensure consistent byte output.
	pbCompaniesPvt := make(map[string]*pb.Company)
	if len(a.CompaniesPvt) > 0 {
		keys := make([]string, 0, len(a.CompaniesPvt))
		for key := range a.CompaniesPvt {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			company := a.CompaniesPvt[key]
			if company != nil {
				pbCompaniesPvt[key] = company.ToProto().(*pb.Company)
			}
		}
	}

	return &pb.AuctionData{
		AuctionID:      a.AuctionID,
		SellBids:       pbSellBids,
		BuyBids:        pbBuyBids,
		ActivePolicies: pbActivePolicies,
		CompaniesPvt:   pbCompaniesPvt,
		Coupled:        a.Coupled,
	}
}

// FromProto implements state.ProtoConvertible.
func (a *AuctionData) FromProto(m proto.Message) error {
	pbAd, ok := m.(*pb.AuctionData)
	if !ok {
		return fmt.Errorf("unexpected proto message type for AuctionData")
	}

	a.AuctionID = pbAd.AuctionID

	a.SellBids = make([]*bids.SellBid, len(pbAd.SellBids))
	for i, pbSellBid := range pbAd.SellBids {
		sellBid := &bids.SellBid{}
		if err := sellBid.FromProto(pbSellBid); err != nil {
			return fmt.Errorf("could not convert SellBid from proto: %v", err)
		}
		a.SellBids[i] = sellBid
	}

	a.BuyBids = make([]*bids.BuyBid, len(pbAd.BuyBids))
	for i, pbBuyBid := range pbAd.BuyBids {
		buyBid := &bids.BuyBid{}
		if err := buyBid.FromProto(pbBuyBid); err != nil {
			return fmt.Errorf("could not convert BuyBid from proto: %v", err)
		}
		a.BuyBids[i] = buyBid
	}

	a.ActivePolicies = make([]policies.Name, len(pbAd.ActivePolicies))
	for i, policyStr := range pbAd.ActivePolicies {
		a.ActivePolicies[i] = policies.Name(policyStr)
	}

	a.CompaniesPvt = make(map[string]*companies.Company)
	for key, pbCompany := range pbAd.CompaniesPvt {
		if pbCompany != nil {
			company := &companies.Company{}
			if err := company.FromProto(pbCompany); err != nil {
				return fmt.Errorf("could not convert Company from proto: %v", err)
			}
			a.CompaniesPvt[key] = company
		}
	}

	a.Coupled = pbAd.Coupled

	return nil
}
