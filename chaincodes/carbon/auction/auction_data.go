package auction

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

type AuctionData struct {
	SellBids       []*bids.SellBid      `json:"sellBids"`
	BuyBids        []*bids.BuyBid       `json:"buyBids"`
	ActivePolicies []policies.Name      `json:"activePolicies"`
	CompaniesPvt   []*companies.Company `json:"buyingCompanies"`
	Coupled        bool                 `json:"coupled"`
}

// TODO: ensure that all data is fetched
func (a *AuctionData) RetrieveData(stub shim.ChaincodeStubInterface, endRFC339Timestamp string) error {
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") != nil {
		return fmt.Errorf("caller does not have the %s attribute, which is required to get prices", identities.PriceViewer)
	}

	var err error
	a.BuyBids, err = state.GetStatesByRangeCompositeKey[bids.BuyBid](stub, bids.BUY_BID_PREFIX, []string{""}, []string{endRFC339Timestamp})
	if err != nil {
		return fmt.Errorf("could not get buy bids: %v", err)
	}

	a.SellBids, err = state.GetStatesByRangeCompositeKey[bids.SellBid](stub, bids.SELL_BID_PREFIX, []string{""}, []string{endRFC339Timestamp})
	if err != nil {
		return fmt.Errorf("could not get sell bids: %v", err)
	}

	for _, buyBid := range a.BuyBids {
		if err := buyBid.FetchPrivatePrice(stub); err != nil {
			return err
		}
	}

	for _, sellBid := range a.SellBids {
		if err := sellBid.FetchPrivatePrice(stub); err != nil {
			return err
		}
		if err := sellBid.FetchCredit(stub); err != nil {
			return err
		}
	}

	var auctionType AuctionType = ""
	err = auctionType.FromWorldState(stub, []string{})
	if err != nil {
		return fmt.Errorf("could not get auction type: %v", err)
	}
	a.Coupled = auctionType == AUCTION_COUPLED

	return err

}

func (a *AuctionData) ToSerializedAuctionData() (*SerializedAuctionData, error) {
	serializedAuctionData := &SerializedAuctionData{}
	var err error

	auctionDataBytes, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("could not marshal auction data: %v", err)
	}

	serializedAuctionData.AuctionDataBytes = auctionDataBytes

	return serializedAuctionData, nil
}
