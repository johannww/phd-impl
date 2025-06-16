package auction

import (
	"bytes"
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
	Sum           []byte // SHA256 sum of bytes of above fields
	Coupled       bool
}

// TODO: ensure that all data is fetched
func (a *AuctionData) RetrieveData(stub shim.ChaincodeStubInterface, endRFC339Timestamp string) error {

	if cid.AssertAttributeValue(stub, identities.PriceViewer, "true") != nil {
		return fmt.Errorf("caller does not have the %s attribute, which is required to get prices", identities.PriceViewer)
	}

	var err error
	buyBids, err := state.GetStatesByRangeCompositeKey[bids.BuyBid](stub, bids.BUY_BID_PREFIX, []string{""}, []string{endRFC339Timestamp})
	if err != nil {
		return fmt.Errorf("could not get buy bids: %v", err)
	}

	sellBids, err := state.GetStatesByRangeCompositeKey[bids.SellBid](stub, bids.SELL_BID_PREFIX, []string{""}, []string{endRFC339Timestamp})
	if err != nil {
		return fmt.Errorf("could not get sell bids: %v", err)
	}

	for _, buyBid := range buyBids {
		if err := buyBid.FetchPrivatePrice(stub); err != nil {
			return err
		}
	}

	for _, sellBid := range sellBids {
		if err := sellBid.FetchPrivatePrice(stub); err != nil {
			return err
		}
		if err := sellBid.FetchCredit(stub); err != nil {
			return err
		}
	}

	err = a.CalculateHash()
	if err != nil {
		return fmt.Errorf("could not calculate sum: %v", err)
	}

	auctionCommitment := &AuctionCommitment{
		EndTimestamp: endRFC339Timestamp,
		Hash:         a.Sum,
	}
	err = auctionCommitment.ToWorldState(stub)

	return err

}

func (a *AuctionData) CalculateHash() error {
	sum, err := a.calculateHash()
	if err != nil {
		return fmt.Errorf("could not calculate auction data hash: %v", err)
	}

	a.Sum = sum
	return nil
}

func (a *AuctionData) calculateHash() ([]byte, error) {
	hash := sha256.New()
	for _, b := range a.BuyBidsBytes {
		_, err := hash.Write(b)
		if err != nil {
			return nil, fmt.Errorf("could not write sell bid bytes to hash: %v", err)
		}
	}

	for _, s := range a.SellBidsBytes {
		_, err := hash.Write(s)
		if err != nil {
			return nil, fmt.Errorf("could not write sell bid bytes to hash: %v", err)
		}
	}

	return hash.Sum(nil), nil
}

func (a *AuctionData) ValidateHash() bool {
	if a.Sum == nil {
		return false
	}

	calculatedSum, err := a.calculateHash()
	return err == nil && bytes.Equal(a.Sum, calculatedSum)
}
