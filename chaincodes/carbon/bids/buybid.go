package bids

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon"
	ccstate "github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	BUY_BID_PREFIX = "buyBid"
	BUY_BID_PVT    = "buyBidPvt"
)

// BuyBid represents an ask from a buyer.
// Their ID could be either x509 or pseudonym-based
type BuyBid struct {
	ID          []string        `json:"id"`
	AskPriceID  PrivatePrice    `json:"askPrice"`
	AskQuantity float64         `json:"askQuantity"`
	BuyerID     carbon.Identity `json:"buyerID"`
}

func PublishBuyBid(stub shim.ChaincodeStubInterface, quantity float64, buyerID carbon.Identity) error {
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

	bidID := []string{buyerID.String(), bidTS.String()}

	privatePrice := &PrivatePrice{
		Price: float64(price),
		BidID: bidID,
	}

	if err := ccstate.PutPvtDataWithCompositeKey[*PrivatePrice](stub, BUY_BID_PVT, bidID, PVT_DATA_COLLECTION, privatePrice); err != nil {
		return err
	}

	buyBid := &BuyBid{
		ID:          bidID,
		AskQuantity: quantity,
		BuyerID:     buyerID,
	}
	if err := ccstate.PutStateWithCompositeKey[*BuyBid](stub, BUY_BID_PREFIX, bidID, buyBid); err != nil {
		return err
	}
	return nil
}

func RetractBuyBid(stub shim.ChaincodeStubInterface, bidID []string) error {
	if err := retractBid(stub, BUY_BID_PREFIX, bidID); err != nil {
		return err
	}
	return nil
}
