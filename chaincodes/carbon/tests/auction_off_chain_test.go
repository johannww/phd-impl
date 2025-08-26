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

func TestOffChainIndependentAuction(t *testing.T) {
	stub, testData := genTestDataAndStub()

	issueStart, err := time.Parse(time.RFC3339, "2023-01-01T00:31:00Z")
	require.NoError(t, err)
	issueEnd := genAllMatchedBids(testData, issueStart)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	setAuctionType(t, stub, auction.AUCTION_INDEPENDENT, "tx2")

	auctionData := retriveAuctionDataFromWorldState(t, stub, issueEnd, testData, "tx3")

	var totalBuyBidQuantity int64 = 0
	for _, bid := range auctionData.BuyBids {
		totalBuyBidQuantity += bid.AskQuantity
	}

	// Run the auction
	auctionResultPub, auctionResultPvt, err := auction.RunIndependent(auctionData)
	require.NoError(t, err, "Failed to run independent auction")
	auctionResult, err := auction.MergeIndependentPublicPrivateResults(auctionResultPub, auctionResultPvt)
	require.NoError(t, err, "Failed to merge independent auction results")

	verifyBidsQuantityConsistency(t, totalBuyBidQuantity, auctionResult)
}

// TODOHP: continue here. test with random bids for credits
func TestOffChainIndependentAuctionWithRandomBids(t *testing.T) {
	stub, testData := genTestDataAndStub()

	issueStart, err := time.Parse(time.RFC3339, "2023-01-01T00:31:00Z")
	require.NoError(t, err)
	issueEnd := genRandomBidsForMintCredits(testData, issueStart)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	setAuctionType(t, stub, auction.AUCTION_INDEPENDENT, "tx2")

	auctionData := retriveAuctionDataFromWorldState(t, stub, issueEnd, testData, "tx3")

	var totalBuyBidQuantity int64 = 0
	for _, bid := range auctionData.BuyBids {
		totalBuyBidQuantity += bid.AskQuantity
	}

	// Run the auction
	auctionResultPub, auctionResultPvt, err := auction.RunIndependent(auctionData)
	require.NoError(t, err, "Failed to run independent auction")
	auctionResult, err := auction.MergeIndependentPublicPrivateResults(auctionResultPub, auctionResultPvt)
	require.NoError(t, err, "Failed to merge independent auction results")

	verifyBidsQuantityConsistency(t, totalBuyBidQuantity, auctionResult)
}

// verifyBidsQuantityConsistency verifies that the sum of quantities of matched bids and adjusted bids equals total bids quantity
func verifyBidsQuantityConsistency(
	t *testing.T,
	totalBuyBidQuantity int64,
	auctionResult *auction.OffChainIndepAuctionResult,
) {
	totalMatched := int64(0)
	totalAdjusted := int64(0)

	for _, bid := range auctionResult.MatchedBids {
		totalMatched += bid.Quantity
	}

	for _, bid := range auctionResult.AdustedBuyBids {
		totalAdjusted += bid.AskQuantity
	}

	require.Equal(t, totalBuyBidQuantity, totalMatched+totalAdjusted, "Sum of matched and adjusted bids quantities must equal total bids quantity")
}


func genTestDataAndStub() (*mocks.MockStub, *utils_test.TestData) {
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
	return stub, testData
}

func setAuctionType(t *testing.T, stub *mocks.MockStub, auctionType auction.AuctionType, txID string) {
	stub.MockTransactionStart(txID)
	err := auctionType.ToWorldState(stub)
	stub.MockTransactionEnd(txID)
	require.NoError(t, err, "Failed to set auction type in world state")
}

func retriveAuctionDataFromWorldState(
	t *testing.T, stub *mocks.MockStub, issueEnd string,
	testData *utils_test.TestData,
	txID string,
) *auction.AuctionData {
	stub.MockTransactionStart(txID)
	stub.Creator = (*testData.Identities)[identities.PriceViewer]
	auctionData := &auction.AuctionData{}
	auctionID, err := auction.IncrementAuctionID(stub)
	require.NoError(t, err, "Failed to increment auction ID")
	err = auctionData.RetrieveData(stub, issueEnd)
	auctionData.AuctionID = auctionID
	stub.MockTransactionEnd(txID)
	require.NoError(t, err, "Failed to retrieve auction data")
	return auctionData
}

func genRandomBidsForMintCredits(testData *utils_test.TestData, issueStart time.Time) (lastIssueRFC339Ts string) {
	sellMinPrice := int64(1000)
	buyMinPrice := int64(1000)

	buyerIds := testData.PseudonymMap

	var issueTsStr string
	for i, mintCredit := range testData.MintCredits {
		issueTs := issueStart.Add(time.Duration(time.Duration(i) * time.Second)).UTC()
		issueTsStr = issueTs.Format(time.RFC3339)
		sellPrice := sellMinPrice + int64(rand.Intn(1000)) // Randomize sell price
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

		buyPrice := buyMinPrice + int64(rand.Intn(1000)) // Randomize buy price
		buyBid := &bids.BuyBid{
			BuyerID:     buyerIds[buyerIdIndex].Pseudonym,
			AskQuantity: mintCredit.Quantity,
			Timestamp:   issueTsStr,
			PrivatePrice: &bids.PrivatePrice{
				Price: buyPrice,
			},
		}
		buyBid.PrivatePrice.BidID = (*buyBid.GetID())[0]

		testData.BuyBids = append(testData.BuyBids, buyBid)
	}

	return issueTsStr
}
