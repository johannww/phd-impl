package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	SELL_BID_PREFIX = "sellBid"
	SELL_BID_PVT    = "sellBidPvt"
)

type SellBid struct {
	CreditID     uint64        `json:"creditID"`
	Timestamp    string        `json:"timestamp"`
	AskQuantity  float64       `json:"askQuantity"`
	PrivatePrice *PrivatePrice `json:"privatePrice"`
}

func PublishSellBid(stub shim.ChaincodeStubInterface, quantity float64, creditID uint64) error {
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

	sellBid := &SellBid{
		CreditID:    creditID,
		Timestamp:   bidTS.String(),
		AskQuantity: quantity,
	}
	bidID := sellBid.GetID()

	privatePrice := &PrivatePrice{
		Price: float64(price),
		BidID: bidID,
	}
	sellBid.PrivatePrice = privatePrice

	if err := sellBid.ToWorldState(stub); err != nil {
		return err
	}

	return nil
}

func RetractSellBid(stub shim.ChaincodeStubInterface, bidID []string) error {
	if err := retractBid(stub, SELL_BID_PREFIX, bidID); err != nil {
		return err
	}
	return nil
}

func (s *SellBid) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := ccstate.GetStateWithCompositeKey(stub, SELL_BID_PREFIX, keyAttributes, s)
	if err != nil {
		return err
	}

	if cid.AssertAttributeValue(stub, "price_viewer", "true") == nil {
		privatePrice := &PrivatePrice{}
		privatePrice.FromWorldState(stub, s.GetID(), SELL_BID_PVT)
		s.PrivatePrice = privatePrice
	}

	return nil
}

func (s *SellBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if s.CreditID == 0 {
		return fmt.Errorf("creditID is not set")
	}
	if s.Timestamp == "" {
		return fmt.Errorf("timestamp is empty")
	}
	if s.AskQuantity <= 0 {
		return fmt.Errorf("askQuantity is not set")
	}

	if s.PrivatePrice != nil {
		err := s.PrivatePrice.ToWorldState(stub, SELL_BID_PVT)
		if err != nil {
			return fmt.Errorf("could not put private price in state: %v", err)
		}
		s.PrivatePrice = nil // Let's not store private data in the world state
	}

	if err := ccstate.PutStateWithCompositeKey(stub, SELL_BID_PREFIX, s.GetID(), s); err != nil {
		return fmt.Errorf("could not put sellbid in state: %v", err)
	}

	return nil
}

func (s *SellBid) GetID() []string {
	// TODO: possible colision with other bids
	return []string{strconv.FormatUint(s.CreditID, 10), s.Timestamp}
}
