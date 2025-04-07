package carbon

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivatePrice is an for-the-government-only price encoded as a base64 string.
type PrivatePrice struct {
	Price float64  `json:"price"`
	BidID []string `json:"bidID"`
}

// BuyBid represents an ask from a buyer.
// Their ID could be either x509 or pseudonym-based
type BuyBid struct {
	ID          []string     `json:"id"`
	AskPriceID  PrivatePrice `json:"askPrice"`
	AskQuantity float64      `json:"askQuantity"`
	BuyerID     Identity     `json:"buyerID"`
}

type SellBid struct {
	ID          []string     `json:"id"`
	AskPriceID  PrivatePrice `json:"askPrice"`
	AskQuantity float64      `json:"askQuantity"`
	CreditID    uint64       `json:"creditID"`
}

const (
	PVT_DATA_COLLECTION = "privateDataCollection"
	BUY_BID_PREFIX      = "buyBid"
	BUY_BID_PVT         = "buyBidPvt"
	SELL_BID_PREFIX     = "sellBid"
	SELL_BID_PVT        = "sellBidPvt"
)

func PublishBuyBid(stub shim.ChaincodeStubInterface, quantity float64, buyerID Identity) error {
	priceBytes, err := getTransientData(stub, "price")
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

	bidID := []string{buyerID.String(), bidTS.String()}

	privatePrice := &PrivatePrice{
		Price: float64(price),
		BidID: bidID,
	}

	if err := putPvtDataWithCompositeKey[*PrivatePrice](stub, BUY_BID_PVT, bidID, PVT_DATA_COLLECTION, privatePrice); err != nil {
		return err
	}

	buyBid := &BuyBid{
		ID:          bidID,
		AskQuantity: quantity,
		BuyerID:     buyerID,
	}
	if err := putStateWithCompositeKey[*BuyBid](stub, BUY_BID_PREFIX, bidID, buyBid); err != nil {
		return err
	}
	return nil
}

func PublisSellBid(stub shim.ChaincodeStubInterface, quantity float64, creditID uint64) error {
	priceBytes, err := getTransientData(stub, "price")
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

	bidID := []string{strconv.FormatUint(creditID, 10), bidTS.String()}

	privatePrice := &PrivatePrice{
		Price: float64(price),
		BidID: bidID,
	}

	if err := putPvtDataWithCompositeKey[*PrivatePrice](stub, SELL_BID_PVT, bidID, PVT_DATA_COLLECTION, privatePrice); err != nil {
		return err
	}

	sellBid := &SellBid{
		ID:          bidID,
		AskQuantity: quantity,
		CreditID:    creditID,
	}
	if err := putStateWithCompositeKey[*SellBid](stub, SELL_BID_PREFIX, bidID, sellBid); err != nil {
		return err
	}
	return nil
}
