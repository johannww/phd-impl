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
	ccstate.WorldStateReconstructor
	CreditID    uint64  `json:"creditID"`
	Timestamp   string  `json:"timestamp"`
	AskQuantity float64 `json:"askQuantity"`
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

	if err := ccstate.PutPvtDataWithCompositeKey[*PrivatePrice](stub, SELL_BID_PVT, bidID, PVT_DATA_COLLECTION, privatePrice); err != nil {
		return err
	}

	if err := ccstate.PutStateWithCompositeKey[*SellBid](stub, SELL_BID_PREFIX, bidID, sellBid); err != nil {
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


func (s *SellBid) GetID() []string {
	return []string{strconv.FormatUint(s.CreditID, 10), s.Timestamp}
}
