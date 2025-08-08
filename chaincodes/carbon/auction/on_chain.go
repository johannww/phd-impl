package auction

import (
	"fmt"
	"slices"
	"sync"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

const (
	PAYMENT_CHAN_SIZE = 100
)

// RunOnChainAuction runs the auction on-chain. This, however, does not provide
// hardware attestation like the off-chain auction. Only the organizations with
// access to the private data can see the bids.
func RunOnChainAuction(stub shim.ChaincodeStubInterface) error {
	err := cid.AssertAttributeValue(stub, identities.PriceViewer, "true")
	if err != nil {
		return fmt.Errorf("caller does not have the %s attribute: %v", identities.PriceViewer, err)
	}

	sellBids, buyBids, err := getBids(stub)
	if err != nil {
		return fmt.Errorf("could not get bids: %v", err)
	}

	if len(sellBids) == 0 {
		return fmt.Errorf("no sell bids found")
	}

	if len(buyBids) == 0 {
		return fmt.Errorf("no buy bids found")
	}

	matchedBids, err := matchBidsIndependent(stub, sellBids, buyBids)
	if err != nil {
		return fmt.Errorf("could not match bids: %v", err)
	}

	_ = matchedBids

	return nil

}

func getBids(stub shim.ChaincodeStubInterface) ([]*bids.SellBid, []*bids.BuyBid, error) {
	var buyBids []*bids.BuyBid
	var sellBids []*bids.SellBid
	var err1, err2 error

	protoTs, _ := stub.GetTxTimestamp()
	ts := utils.TimestampRFC3339UtcString(protoTs)
	rangeStart, rangeEnd := []string{""}, []string{ts}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		buyBids, err1 = state.GetStatesByRangeCompositeKey[bids.BuyBid](
			stub,
			bids.BUY_BID_PREFIX,
			rangeStart, rangeEnd)

		for _, bid := range buyBids {
			if err := bid.FetchPrivatePrice(stub); err != nil {
				err1 = fmt.Errorf("could not fetch private price for sell bid: %v", err)
				break
			}
		}
		wg.Done()
	}()

	go func() {
		sellBids, err2 = state.GetStatesByRangeCompositeKey[bids.SellBid](
			stub,
			bids.SELL_BID_PREFIX,
			rangeStart, rangeEnd)
		for _, bid := range sellBids {
			if err := bid.FetchPrivatePrice(stub); err != nil {
				err2 = fmt.Errorf("could not fetch private price for sell bid: %v", err)
				break
			}
		}
		wg.Done()
	}()

	wg.Wait()

	if err1 != nil {
		return nil, nil, fmt.Errorf("could not get buy bids batch: %v", err1)
	}
	if err2 != nil {
		return nil, nil, fmt.Errorf("could not get sell bids batch: %v", err2)
	}

	return sellBids, buyBids, nil
}

func matchBidsIndependent(
	stub shim.ChaincodeStubInterface,
	sellBids []*bids.SellBid,
	buyBids []*bids.BuyBid,
) ([]*bids.MatchedBid, error) {
	slices.SortFunc(sellBids, func(a, b *bids.SellBid) int {
		return a.Less(b)
	})
	slices.SortFunc(buyBids, func(a, b *bids.BuyBid) int {
		return a.Less(b)
	})

	creditChan := make(chan *bids.MatchedBid, PAYMENT_CHAN_SIZE)
	paymentTokenChan := make(chan *bids.MatchedBid, PAYMENT_CHAN_SIZE)
	defer close(creditChan)
	defer close(paymentTokenChan)

	go transferCredits(stub, creditChan)

	matchedBids := make([]*bids.MatchedBid, 0)

	i, j := 0, len(buyBids)-1
	lastMatch := [2]int{-1, -1}
	hasCuttingPrice := buyBids[j].PrivatePrice.Price < sellBids[i].PrivatePrice.Price
	for buyBids[j].PrivatePrice.Price > sellBids[i].PrivatePrice.Price {
		matchQuantity := min(sellBids[i].Quantity, buyBids[j].AskQuantity)
		sellBids[i].Quantity -= matchQuantity
		buyBids[j].AskQuantity -= matchQuantity

		matchedBid := &bids.MatchedBid{
			BuyBidID:  (*buyBids[j].GetID())[0],
			BuyBid:    buyBids[j],
			SellBidID: (*sellBids[i].GetID())[0],
			SellBid:   sellBids[i],
			Quantity:  matchQuantity,
		}

		creditChan <- matchedBid

		matchedBids = append(matchedBids, matchedBid)

		lastMatch[0] = i
		lastMatch[1] = j

		var err error
		if sellBids[i].Quantity == 0 {
			err = sellBids[i].DeleteFromWorldState(stub)
			i++
		} else {
			err = buyBids[j].DeleteFromWorldState(stub)
			j--
		}

		if err != nil {
			return nil, fmt.Errorf("could not delete bid from world state: %v", err)
		}

		if i >= len(buyBids) || j < 0 {
			break
		}
	}

	if !hasCuttingPrice {
		return nil, fmt.Errorf("no cutting price found")
	}

	cuttingPrice := buyBids[lastMatch[0]].PrivatePrice.Price + sellBids[lastMatch[1]].PrivatePrice.Price/2

	go transferPaymentToken(stub, paymentTokenChan)

	for _, matchedBid := range matchedBids {
		matchedBid.PrivatePrice.Price = cuttingPrice
		matchedBid.PrivatePrice.BidID = (*matchedBid.GetID())[0]
		paymentTokenChan <- matchedBid
		err := matchedBid.ToWorldState(stub)
		if err != nil {
			return nil, fmt.Errorf("could not put matched bid in world state: %v", err)
		}
	}

	return nil, nil
}

func transferCredits(stub shim.ChaincodeStubInterface, matchedChan chan *bids.MatchedBid) {
	creditWalletCache := map[string]*credits.CreditWallet{}

	for matchedBid := range matchedChan {

		// get credit wallet
		buyerCW, err := state.GetStateFromCache(stub, &creditWalletCache, credits.CREDIT_WALLET_PREFIX, []string{matchedBid.BuyBid.BuyerID})
		if err != nil {
			// close(matchedChan)
			panic(fmt.Sprintf("could not get buyer %s credit wallet: %s", matchedBid.BuyBid.BuyerID, err.Error()))
		}
		sellerCW, err := state.GetStateFromCache(stub, &creditWalletCache, credits.CREDIT_WALLET_PREFIX, []string{matchedBid.SellBid.SellerID})
		if err != nil {
			close(matchedChan)
			panic("could not get seller credit wallet: " + err.Error())
		}

		buyerCW.Quantity += matchedBid.Quantity
		sellerCW.Quantity -= matchedBid.Quantity
	}

	// Save the cached credit wallets
	for _, cw := range creditWalletCache {
		err := cw.ToWorldState(stub)
		if err != nil {
			close(matchedChan)
			panic("could not put credit wallet in world state: " + err.Error())
		}
	}
}

func transferPaymentToken(stub shim.ChaincodeStubInterface, matchedChan chan *bids.MatchedBid) {
	vtwCache := map[string]*payment.VirtualTokenWallet{}

	for matchedBid := range matchedChan {

		//get virtual token wallet
		buyerVtw, err := state.GetStateFromCache(stub, &vtwCache, payment.VIRTUAL_TOKEN_WALLET_PREFIX, []string{matchedBid.BuyBid.BuyerID})
		if err != nil {
			// close(matchedChan)
			panic("could not get buyer virtual token wallet: " + err.Error())
		}
		sellerVtw, err := state.GetStateFromCache(stub, &vtwCache, payment.VIRTUAL_TOKEN_WALLET_PREFIX, []string{matchedBid.SellBid.SellerID})
		if err != nil {
			// close(matchedChan)
			panic("could not get seller virtual token wallet: " + err.Error())
		}

		buyerVtw.Quantity -= matchedBid.PrivatePrice.Price * matchedBid.Quantity
		sellerVtw.Quantity += matchedBid.PrivatePrice.Price * matchedBid.Quantity

	}

	// Save the cached payment token wallets
	for _, vtw := range vtwCache {
		err := vtw.ToWorldState(stub)
		if err != nil {
			// close(matchedChan)
			return
		}
	}
}
