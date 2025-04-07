package carbon

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// TODO: this may be a float64 passed to the chaincode via transient data
// PrivatePrice is an for-the-government-only price encoded as a base64 string.
type PrivatePrice string

// BuyBid represents an ask from a buyer.
// Their ID could be either x509 or pseudonym-based
type BuyBid struct {
	AskPrice    PrivatePrice `json:"askPrice"`
	AskQuantity float64      `json:"askQuantity"`
	BuyerID     Identity     `json:"buyerID"`
}

type SellBid struct {
	AskPrice    PrivatePrice `json:"askPrice"`
	AskQuantity float64      `json:"askQuantity"`
	PropertyID  uint64       `json:"propertyID"`
}

const (
	PVT_DATA_COLLECTION = "privateDataCollection"
	BUY_BID_PREFIX      = "buyBid"
	SELL_BID_PREFIX     = "sellBid"
)

func PublishBuyBid(quantity float64, buyerID Identity, stub shim.ChaincodeStubInterface) error {
	buyBid := &BuyBid{
		AskPrice:    price,
		AskQuantity: quantity,
		BuyerID:     buyerID,
	}
	price, err := getTransientData(stub, "price")
	if err != nil {
		return err
	}

	stub.PutPrivateData(PVT_DATA_COLLECTION, "buyBid", []byte(buyBid.AskPrice))
	bidKey := stub.CreateCompositeKey(BUY_BID_PREFIX, []string{string(buyBid.AskPrice), string(buyBid.AskQuantity)})
	// mocks.ChaincodeStub.
	return buyBid
}
