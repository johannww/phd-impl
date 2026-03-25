package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	ccstate "github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
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

func (s *SellBid) canReadPrivateBidData(stub shim.ChaincodeStubInterface) bool {
	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") == nil {
		return true
	}
	return identities.GetID(stub) == s.SellerID
}

func (s *SellBid) FetchPrivatePrice(stub shim.ChaincodeStubInterface) error {
	if s.canReadPrivateBidData(stub) {
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

	// Only sell bids from credits.MintCredit have associated credit data to fetch.
	if len(s.CreditID) > 0 {
		err = s.FetchCredit(stub)
		if err != nil {
			return err
		}
	}

	err = s.FetchPrivatePrice(stub)
	if err != nil {
		return err
	}

	return nil
}

// TODO: test for the bids mutex timestamp
func (s *SellBid) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if s.Timestamp == "" {
		return fmt.Errorf("timestamp is empty")
	}
	if s.Quantity <= 0 {
		return fmt.Errorf("Quantity is not set")
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
	if err = ccstate.PutStateWithCompositeKey(stub, SELL_BID_PREFIX, copyS.GetID(), &copyS); err != nil {
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

func (s *SellBid) Less(b2 *SellBid) int {
	if s.PrivatePrice.Price < b2.PrivatePrice.Price {
		return -1
	} else if s.PrivatePrice.Price > b2.PrivatePrice.Price {
		return 1
	}
	return 0
}

func (s *SellBid) DeepCopy() *SellBid {
	bidCopy := *s
	if s.Credit != nil {
		creditCopy := *s.Credit
		bidCopy.Credit = &creditCopy
	}
	if s.PrivatePrice != nil {
		priceCopy := *s.PrivatePrice
		bidCopy.PrivatePrice = &priceCopy
	}
	return &bidCopy
}

func (s *SellBid) ToProto() proto.Message {
	var pbCredit *pb.MintCredit
	if s.Credit != nil {
		pbCredit = s.Credit.ToProto().(*pb.MintCredit)
	}

	var pbPrivPrice *pb.PrivatePrice
	if s.PrivatePrice != nil {
		pbPrivPrice = s.PrivatePrice.ToProto().(*pb.PrivatePrice)
	}

	return &pb.SellBid{
		SellerID:     s.SellerID,
		CreditID:     s.CreditID,
		Timestamp:    s.Timestamp,
		Credit:       pbCredit,
		Quantity:     s.Quantity,
		PrivatePrice: pbPrivPrice,
	}
}

func (s *SellBid) FromProto(m proto.Message) error {
	pbSell, ok := m.(*pb.SellBid)
	if !ok {
		return fmt.Errorf("unexpected proto message type for SellBid")
	}
	s.SellerID = pbSell.SellerID
	s.CreditID = pbSell.CreditID
	s.Timestamp = pbSell.Timestamp
	s.Quantity = pbSell.Quantity

	if pbSell.PrivatePrice != nil {
		pp := &PrivatePrice{}
		pp.Price = pbSell.PrivatePrice.Price
		pp.BidID = pbSell.PrivatePrice.BidID
		s.PrivatePrice = pp
	}

	if pbSell.Credit != nil {
		mc := &credits.MintCredit{}
		err := mc.FromProto(pbSell.Credit)
		if err != nil {
			return fmt.Errorf("could not map pb.MintCredit to credits.MintCredit: %v", err)
		}
		s.Credit = mc
	}

	return nil
}

// PublishSellBidFromCredit creates a sell bid in the world state from a mint credit.
// The price is read from the transient data.
func PublishSellBidFromCredit(stub shim.ChaincodeStubInterface, quantity int64, creditID []string) error {
	ownerID, price, bidTSStr, err := validateAndExtractSellBidInput(stub, quantity)
	if err != nil {
		return err
	}

	sellBid := &SellBid{
		SellerID:  ownerID,
		CreditID:  creditID,
		Timestamp: bidTSStr,
		Quantity:  quantity,
	}

	if err := createSellBidFromCredit(stub, sellBid); err != nil {
		return fmt.Errorf("could not create sell bid from mint credit: %v", err)
	}

	sellBid.PrivatePrice = &PrivatePrice{
		Price: price,
		BidID: (*sellBid.GetID())[0],
	}

	if err := sellBid.ToWorldState(stub); err != nil {
		return err
	}

	return nil
}

// PublishSellBidFromWallet creates a sell bid by taking credits from the seller's
// fungible credit wallet (CreditWallet).
func PublishSellBidFromWallet(stub shim.ChaincodeStubInterface, quantity int64) error {
	ownerID, price, bidTSStr, err := validateAndExtractSellBidInput(stub, quantity)
	if err != nil {
		return err
	}

	sellBid := &SellBid{
		SellerID:  ownerID,
		Timestamp: bidTSStr,
		Quantity:  quantity,
	}

	if err := createSellBidFromWallet(stub, sellBid); err != nil {
		return fmt.Errorf("could not create sell bid from credit wallet: %v", err)
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

// CreateSellBidFromWallet moves credits from the mint credit to the seller's credit wallet.
func createSellBidFromCredit(stub shim.ChaincodeStubInterface, s *SellBid) error {
	if err := s.FetchCredit(stub); err != nil {
		return fmt.Errorf("could not fetch credit: %v", err)
	}

	if s.Credit.Quantity < s.Quantity {
		return fmt.Errorf("sell bid quantity %d exceeds mint credit quantity %d", s.Quantity, s.Credit.Quantity)
	}

	// Decrease mint credit quantity
	s.Credit.Quantity -= s.Quantity

	// Persist updated mint credit (we remove quantity from the mint)
	if err := s.Credit.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not put updated mint credit in state: %v", err)
	}

	return nil
}

func RetractSellBid(stub shim.ChaincodeStubInterface, bidID []string) error {
	if len(bidID) < 2 {
		return fmt.Errorf("invalid bid ID: expected at least 2 attributes")
	}

	sellBid := &SellBid{
		Timestamp: bidID[0],
		SellerID:  bidID[1],
	}

	callerID := identities.GetID(stub)
	if callerID != sellBid.SellerID {
		return fmt.Errorf("caller is not the bid owner")
	}

	if err := sellBid.FromWorldState(stub, bidID); err != nil {
		return fmt.Errorf("could not get sell bid from world state: %v", err)
	}

	if len(sellBid.CreditID) == 0 {
		return retractWalletBasedSellBid(stub, sellBid)
	}

	return retractMintCreditBasedSellBid(stub, sellBid)

}

// CreateSellBidFromTokenWallet deducts the sell quantity from the seller's
// fungible CreditWallet and persists the wallet.
func createSellBidFromWallet(stub shim.ChaincodeStubInterface, s *SellBid) error {
	sellerCW := &credits.CreditWallet{OwnerID: s.SellerID}
	if err := sellerCW.FromWorldState(stub, (*sellerCW.GetID())[0]); err != nil {
		return fmt.Errorf("could not get seller credit wallet: %v", err)
	}

	if sellerCW.Quantity < s.Quantity {
		return fmt.Errorf("seller credit wallet has insufficient quantity: have %d, need %d", sellerCW.Quantity, s.Quantity)
	}

	sellerCW.Quantity -= s.Quantity

	if err := sellerCW.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not update seller credit wallet: %v", err)
	}

	s.Credit = nil // mark that there is no associated MintCredit

	return nil
}

func validateAndExtractSellBidInput(stub shim.ChaincodeStubInterface, quantity int64) (string, int64, string, error) {
	if quantity <= 0 {
		return "", 0, "", fmt.Errorf("quantity must be greater than zero")
	}

	ownerID := identities.GetID(stub)

	priceBytes, err := ccstate.GetTransientData(stub, "price")
	if err != nil {
		return "", 0, "", fmt.Errorf("could not get price from transient data: %v", err)
	}

	price, err := strconv.ParseInt(string(priceBytes), 10, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("could not parse price: %v", err)
	}

	bidTS, err := stub.GetTxTimestamp()
	if err != nil {
		return "", 0, "", fmt.Errorf("could not get transaction timestamp: %v", err)
	}
	bidTSStr := utils.TimestampRFC3339UtcString(bidTS)

	return ownerID, price, bidTSStr, nil
}

func retractWalletBasedSellBid(stub shim.ChaincodeStubInterface, sellBid *SellBid) error {
	sellerCW := &credits.CreditWallet{OwnerID: sellBid.SellerID}
	if err := sellerCW.FromWorldState(stub, []string{sellBid.SellerID}); err != nil {
		return fmt.Errorf("could not get seller credit wallet: %v", err)
	}

	sellerCW.Quantity += sellBid.Quantity

	if err := sellerCW.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not update seller credit wallet: %v", err)
	}

	return sellBid.DeleteFromWorldState(stub)
}

func retractMintCreditBasedSellBid(stub shim.ChaincodeStubInterface, sellBid *SellBid) error {
	if err := sellBid.FetchCredit(stub); err != nil {
		return fmt.Errorf("could not fetch credit for sell bid: %v", err)
	}

	sellBid.Credit.Quantity += sellBid.Quantity

	if err := sellBid.Credit.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not update mint credit in world state: %v", err)
	}

	return sellBid.DeleteFromWorldState(stub)
}
