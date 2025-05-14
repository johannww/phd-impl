package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

const (
	SELL_BID_PREFIX = "sellBid"
	SELL_BID_PVT    = "sellBidPvt"
)

// TODO: review how the credit should be loaded here
type SellBid struct {
	CreditID     string          `json:"creditID"`
	Timestamp    string          `json:"timestamp"`
	Credit       *credits.Credit `json:"credit"`
	AskQuantity  float64         `json:"askQuantity"`
	PrivatePrice *PrivatePrice   `json:"-"`
}

var _ ccstate.WorldStateManager = (*SellBid)(nil)

func PublishSellBid(stub shim.ChaincodeStubInterface, quantity float64, creditID string) error {
	priceBytes, err := ccstate.GetTransientData(stub, "price")
	if err != nil {
		return err
	}

	price, err := strconv.ParseFloat(string(priceBytes), 64)
	if err != nil {
		return fmt.Errorf("could not parse price: %v", err)
	}

	bidTS, err := stub.GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("could not get transaction timestamp: %v", err)
	}
	bidTSStr := utils.TimestampRFC3339UtcString(bidTS)

	sellBid := &SellBid{
		CreditID:    creditID,
		Timestamp:   bidTSStr,
		AskQuantity: quantity,
	}
	bidID := *(sellBid.GetID())

	privatePrice := &PrivatePrice{
		Price: float64(price),
		BidID: bidID[0],
	}
	sellBid.PrivatePrice = privatePrice

	if err := sellBid.ToWorldState(stub); err != nil {
		return err
	}

	return nil
}

func RetractSellBid(stub shim.ChaincodeStubInterface, bidID []string) error {
	mockBid := &SellBid{
		CreditID:  bidID[0],
		Timestamp: bidID[1],
	}
	err := mockBid.DeleteFromWorldState(stub)
	return err
}

func (s *SellBid) FetchPrivatePrice(stub shim.ChaincodeStubInterface) error {
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") == nil {
		privatePrice := &PrivatePrice{}
		err := privatePrice.FromWorldState(stub, (*s.GetID())[0], SELL_BID_PVT)
		if err != nil {
			return fmt.Errorf("could not get private price from world state: %v", err)
		}
		s.PrivatePrice = privatePrice
	}
	return nil
}

func (s *SellBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := ccstate.GetStateWithCompositeKey(stub, SELL_BID_PREFIX, keyAttributes, s)
	if err != nil {
		return err
	}

	// TODO: load credit from world state.
	// perhaps, check if it should be done

	err = s.FetchPrivatePrice(stub)
	if err != nil {
		return err
	}

	return nil
}

// TODO: test for the bids mutex timestamp
func (s *SellBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if s.CreditID == "" {
		return fmt.Errorf("creditID is not set")
	}
	if s.Timestamp == "" {
		return fmt.Errorf("timestamp is empty")
	}
	if s.AskQuantity <= 0 {
		return fmt.Errorf("askQuantity is not set")
	}

	if s.Credit != nil {
		err := s.Credit.ToWorldState(stub)
		if err != nil {
			return fmt.Errorf("could not put credit in state: %v", err)
		}
	}

	if s.PrivatePrice != nil {
		err := s.PrivatePrice.ToWorldState(stub, SELL_BID_PVT)
		if err != nil {
			return fmt.Errorf("could not put private price in state: %v", err)
		}
	}

	if err := ccstate.PutStateWithCompositeKey(stub, SELL_BID_PREFIX, s.GetID(), s); err != nil {
		return fmt.Errorf("could not put sellbid in state: %v", err)
	}

	return nil
}

func (s *SellBid) DeleteFromWorldState(stub shim.ChaincodeStubInterface) error {
	bidID := s.GetID()
	err := ccstate.DeleteStateWithCompositeKey(stub, BUY_BID_PREFIX, bidID)
	if err != nil {
		return fmt.Errorf("could not delete sel bid: %v", err)
	}

	s.PrivatePrice.BidID = (*bidID)[0]
	err = s.PrivatePrice.DeleteFromWorldState(stub, SELL_BID_PVT)

	return err
}

func (s *SellBid) GetID() *[][]string {
	// TODO: possible colision with other bids
	return &[][]string{
		{s.CreditID, s.Timestamp},
	}
}

func (b *SellBid) Less(b2 *SellBid) int {
	if b.PrivatePrice.Price < b2.PrivatePrice.Price {
		return -1
	} else if b.PrivatePrice.Price > b2.PrivatePrice.Price {
		return 1
	}
	return 0
}
