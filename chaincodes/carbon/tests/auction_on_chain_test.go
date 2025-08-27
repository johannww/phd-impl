package carbon_tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/stretchr/testify/require"
)

func TestOnChainIndependentAuction(t *testing.T) {
	nOwners := 10
	nChunks := 3
	nCompanies := 5
	startTimestamp := "2023-01-01T00:00:00Z"
	endTimestamp := "2023-01-01T00:30:00Z"
	issueInterval := 30 * time.Second
	testData := utils_test.GenData(
		nOwners, nChunks, nCompanies,
		startTimestamp, endTimestamp, issueInterval,
	)
	stub := mocks.NewMockStub("carbon", nil)

	issueStart, err := time.Parse(time.RFC3339, "2023-01-01T00:31:00Z")
	require.NoError(t, err)
	genAllMatchedBids(testData, issueStart, auction.AUCTION_INDEPENDENT)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	// t.Fail()
	stub.MockTransactionStart("tx2")
	stub.Creator = (*testData.Identities)[identities.PriceViewer]
	err = auction.RunOnChainAuction(stub)
	stub.MockTransactionEnd("tx2")
	require.NoError(t, err, "RunOnChainAuction should not return an error")

	stub.MockTransactionStart("tx3")
	mintCreditsEqualMatchedCredits(t, stub, testData.MintCredits)
	stub.MockTransactionEnd("tx3")

	stub.MockTransactionStart("tx4")
	buyBidsAndSellBidsAreDeletedFromWorldState(t, stub, testData.BuyBids, testData.SellBids)
	stub.MockTransactionEnd("tx4")
}

// genAllMatchedBids generates a set of bids that will be fully matched
func genAllMatchedBids(testData *utils_test.TestData, issueStart time.Time,
	auctionType auction.AuctionType) (lastIssueRFC339Ts string) {
	sellPrice := int64(1000)
	buyPrice := int64(1200)

	buyerIds := testData.PseudonymMap

	var issueTsStr string
	for i, mintCredit := range testData.MintCredits {
		issueTs := issueStart.Add(time.Duration(time.Duration(i) * time.Second)).UTC()
		issueTsStr = issueTs.Format(time.RFC3339)
		sellBid := &bids.SellBid{
			SellerID:  mintCredit.OwnerID,
			CreditID:  (*mintCredit.GetID())[0],
			Timestamp: issueTsStr,
			PrivatePrice: &bids.PrivatePrice{
				Price: sellPrice,
			},
			Quantity: mintCredit.Quantity,
		}
		sellBid.PrivatePrice.BidID = (*sellBid.GetID())[0]

		testData.SellBids = append(testData.SellBids, sellBid)

		buyerIdIndex := rand.Intn(len(buyerIds))
		buyBid := &bids.BuyBid{
			BuyerID:   buyerIds[buyerIdIndex].Pseudonym,
			Timestamp: issueTsStr,
			PrivatePrice: &bids.PrivatePrice{
				Price: buyPrice,
			},
		}
		if auctionType == auction.AUCTION_COUPLED {
			buyBid.PrivateQuantity = &bids.PrivateQuantity{
				AskQuantity: mintCredit.Quantity,
				BidID:       (*buyBid.GetID())[0],
			}
		} else {
			buyBidAskQuantity := mintCredit.Quantity
			buyBid.AskQuantity = &buyBidAskQuantity
		}
		buyBid.PrivatePrice.BidID = (*buyBid.GetID())[0]

		testData.BuyBids = append(testData.BuyBids, buyBid)
	}

	return issueTsStr
}

// mintCreditsEqualMatchedCredits checks that there are no bids in the world state.
// They should all be deleted after their matched bids are created.
func mintCreditsEqualMatchedCredits(t *testing.T, stub *mocks.MockStub, mintCredits []*credits.MintCredit) {
	mintCreditsTotal := int64(0)
	mintCreditsFromBids := int64(0)

	for _, mintCredit := range mintCredits {
		mintCreditsTotal += mintCredit.Quantity
	}

	matchedBids, err := state.GetStatesByPartialCompositeKey[bids.MatchedBid](stub, bids.MATCHED_BID_PREFIX, nil)
	require.NoError(t, err, "GetStatesByPartialCompositeKey should not return an error")

	for _, matchedBid := range matchedBids {
		mintCreditsFromBids += matchedBid.Quantity
	}
	require.Equal(t, mintCreditsTotal, mintCreditsFromBids, "Mint credits total should be equal to the sum of all bids' quantities")
}

func buyBidsAndSellBidsAreDeletedFromWorldState(t *testing.T, stub *mocks.MockStub, buyBid []*bids.BuyBid, sellBid []*bids.SellBid) {
	buyBids, err := state.GetStatesByPartialCompositeKey[bids.BuyBid](stub, bids.BUY_BID_PREFIX, nil)
	require.NoError(t, err, "GetStatesByPartialCompositeKey should not return an error")
	sellBids, err := state.GetStatesByPartialCompositeKey[bids.SellBid](stub, bids.SELL_BID_PREFIX, nil)
	require.NoError(t, err, "GetStatesByPartialCompositeKey should not return an error")

	for _, bid := range buyBids {
		t.Logf("BuyBid: %+v\n", *bid)
	}
	for _, bid := range sellBids {
		t.Logf("SellBid: %+v\n", *bid)
	}

	require.Len(t, buyBids, 0, "Buy bids should be deleted from the world state")
	require.Len(t, sellBids, 0, "Sell bids should be deleted from the world state")

}
