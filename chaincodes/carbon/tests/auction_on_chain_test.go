package carbon_tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/stretchr/testify/require"
)

// TODO: Finish the test
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
	genAllMatchedBids(testData, issueStart)

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
}

// genAllMatchedBids generates a set of bids that will be fully matched
func genAllMatchedBids(testData *utils_test.TestData, issueStart time.Time) {
	sellPrice := int64(1000)
	buyPrice := int64(1200)

	buyerIds := testData.CompaniesIdentities()

	for i, mintCredit := range testData.MintCredits {
		issueTs := issueStart.Add(time.Duration(time.Duration(i) * time.Second)).UTC()
		issueTsStr := issueTs.Format(time.RFC3339)
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
			BuyerID:     buyerIds[buyerIdIndex],
			AskQuantity: mintCredit.Quantity,
			Timestamp:   issueTsStr,
			PrivatePrice: &bids.PrivatePrice{
				Price: buyPrice,
			},
		}
		buyBid.PrivatePrice.BidID = (*buyBid.GetID())[0]

		testData.BuyBids = append(testData.BuyBids, buyBid)
	}

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
