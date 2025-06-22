package carbon_tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/stretchr/testify/require"
)

// TODO: Finish the test
func TestOnChainIndependentAuction(t *testing.T) {
	testData := utils_test.GenData(
		10, 3, 5,
		"2023-01-01T00:00:00Z",
		"2023-01-01T00:30:00Z", 30*time.Second)
	stub := mocks.NewMockStub("carbon", nil)

	issueStart, err := time.Parse(time.RFC3339, "2023-01-01T00:31:00Z")
	require.NoError(t, err)
	genAllMatchedBids(testData, issueStart)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	// t.Fail()

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
