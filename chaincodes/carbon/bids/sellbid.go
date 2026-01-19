package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

const (
	SELL_BID_PREFIX              = "sellBid"
	SELL_BID_PVT                 = "sellBidPvt"
	SELL_BID_ID_SELLER_AS_PREFIX = 1
	SELL_BID_ID_CREDIT_AS_PREFIX = 2
)

type SellBid struct {
	SellerID     string              `json:"sellerID,omitempty"`
	CreditID     []string            `json:"creditID"`
	Timestamp    string              `json:"timestamp,omitempty"`
	Credit       *credits.MintCredit `json:"credit,omitempty"`
	Quantity     int64               `json:"quantity,omitempty"`
	PrivatePrice *PrivatePrice       `json:"privatePrice,omitempty"`
}

var _ ccstate.WorldStateManager = (*SellBid)(nil)

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

func (s *SellBid) FetchCredit(stub shim.ChaincodeStubInterface) error {
	s.Credit = &credits.MintCredit{}
	err := s.Credit.FromWorldState(stub, s.CreditID)
	if err != nil {
		return fmt.Errorf("could not get credit in state: %v", err)
	}

	if s.Credit.OwnerID != s.SellerID {
		return fmt.Errorf("credit owner ID does not match seller %s", s.SellerID)
	}
	return nil
}

func (s *SellBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := ccstate.GetStateWithCompositeKey(stub, SELL_BID_PREFIX, keyAttributes, s)
	if err != nil {
		return err
	}

	err = s.FetchCredit(stub)
	if err != nil {
		return err
	}

	err = s.FetchPrivatePrice(stub)
	if err != nil {
		return err
	}

	return nil
}

// TODO: test for the bids mutex timestamp
func (s *SellBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if len(s.CreditID) == 0 {
		return fmt.Errorf("creditID is not set")
	}
	if s.Timestamp == "" {
		return fmt.Errorf("timestamp is empty")
	}
	if s.Quantity <= 0 {
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

	copyS := *s              // Create a copy of SellBid to avoid modifying the original
	copyS.PrivatePrice = nil // already stored in the private world state

	var err error
	if err = ccstate.PutStateWithCompositeKey(stub, SELL_BID_PREFIX, copyS.GetID(), copyS); err != nil {
		err = fmt.Errorf("could not put sellbid in state: %v", err)
	}

	return err
}

func (s *SellBid) DeleteFromWorldState(stub shim.ChaincodeStubInterface) error {
	bidID := s.GetID()
	err := ccstate.DeleteStateWithCompositeKey(stub, SELL_BID_PREFIX, bidID)
	if err != nil {
		return fmt.Errorf("could not delete sel bid: %v", err)
	}

	s.PrivatePrice.BidID = (*bidID)[0]
	err = s.PrivatePrice.DeleteFromWorldState(stub, SELL_BID_PVT)

	return err
}

// GetID might result in conflicting IDs if two sell bids are created by the
// same seller for the same credit at the same timestamp. This is unlikely, but should be
// clarified
func (s *SellBid) GetID() *[][]string {
	creditIdAsPrefix := s.CreditID
	creditIdAsPrefix = append(creditIdAsPrefix, s.Timestamp)
	return &[][]string{
		{s.Timestamp, s.SellerID},
		{s.SellerID, s.Timestamp},
		creditIdAsPrefix,
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

func (b *SellBid) DeepCopy() *SellBid {
	bidCopy := *b
	if b.Credit != nil {
		creditCopy := *b.Credit
		bidCopy.Credit = &creditCopy
	}
	if b.PrivatePrice != nil {
		priceCopy := *b.PrivatePrice
		bidCopy.PrivatePrice = &priceCopy
	}
	return &bidCopy
}

// PublishSellBid creates a sell bid in the world state.
// The price is read from the transient data.
// The credit is fetched from the world state to verify ownership and quantity. After that, the credit is updated in the world state with the new quantity.
func PublishSellBid(stub shim.ChaincodeStubInterface, quantity int64, creditID []string) error {
	ownerID := identities.GetID(stub)
	priceBytes, err := ccstate.GetTransientData(stub, "price")
	if err != nil {
		return err
	}

	price, err := strconv.ParseInt(string(priceBytes), 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse price: %v", err)
	}

	bidTS, err := stub.GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("could not get transaction timestamp: %v", err)
	}
	bidTSStr := utils.TimestampRFC3339UtcString(bidTS)

	sellBid := &SellBid{
		SellerID:  ownerID,
		CreditID:  creditID,
		Timestamp: bidTSStr,
		Quantity:  quantity,
	}
	err = sellBid.FetchCredit(stub)
	if err != nil {
		return fmt.Errorf("could not fetch credit: %v", err)
	}

	sellBid.Credit.Quantity -= quantity
	if sellBid.Credit.Quantity < 0 {
		return fmt.Errorf("sell bid quantity %d exceeds credit quantity %d", sellBid.Quantity, sellBid.Credit.Quantity)
	}

	bidID := *(sellBid.GetID())

	privatePrice := &PrivatePrice{
		Price: price,
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
		CreditID:  bidID[0:2],
		Timestamp: bidID[2],
	}
	err := mockBid.DeleteFromWorldState(stub)
	return err
}
