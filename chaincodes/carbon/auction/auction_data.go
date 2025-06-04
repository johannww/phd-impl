package auction

import (
	"crypto/sha256"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

type AuctionData struct {
	SellBidsBytes [][]byte
	BuyBidsBytes  [][]byte
	Sum           []byte // SHA256 sum of bytes
}

// TODO: ensure that all data is fetched
func (a *AuctionData) RetrieveData(stub shim.ChaincodeStubInterface, endRFC339Timestamp string) error {

	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") != nil {
		return fmt.Errorf("caller does not have the %s attribute, which is required to get prices", identities.PriceViewer)
	}

	var err error
	a.BuyBidsBytes, err = state.GetStatesBytesByRangeCompositeKey(stub, bids.BUY_BID_PREFIX, []string{""}, []string{endRFC339Timestamp})
	if err != nil {
		return fmt.Errorf("could not get buy bids: %v", err)
	}

	a.SellBidsBytes, err = state.GetStatesBytesByRangeCompositeKey(stub, bids.SELL_BID_PREFIX, []string{""}, []string{endRFC339Timestamp})
	if err != nil {
		return fmt.Errorf("could not get sell bids: %v", err)
	}

	err = a.CalculateHash()
	if err != nil {
		return fmt.Errorf("could not calculate sum: %v", err)
	}

	return nil

}

func (a *AuctionData) CalculateHash() error {
	hash := sha256.New()
	for _, b := range a.BuyBidsBytes {
		_, err := hash.Write(b)
		if err != nil {
			return fmt.Errorf("could not write sell bid bytes to hash: %v", err)
		}
	}

	for _, s := range a.SellBidsBytes {
		_, err := hash.Write(s)
		if err != nil {
			return fmt.Errorf("could not write sell bid bytes to hash: %v", err)
		}
	}

	a.Sum = hash.Sum(nil)
	return nil
}
