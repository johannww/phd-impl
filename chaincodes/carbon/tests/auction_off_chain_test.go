package carbon_tests

import (
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/stretchr/testify/require"
)

// TODOHP: Continue testing the off-chain independent auction
func TestOffChainIndependentAuction(t *testing.T) {
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
	issueEnd := genAllMatchedBids(testData, issueStart)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	// Set AuctionType
	stub.MockTransactionStart("tx11")
	var auctionType auction.AuctionType = auction.AUCTION_INDEPENDENT
	err = auctionType.ToWorldState(stub)
	stub.MockTransactionEnd("tx11")
	require.NoError(t, err, "Failed to set auction type in world state")

	// Get off-chain auction data from world state
	stub.MockTransactionStart("tx2")
	stub.Creator = (*testData.Identities)[identities.PriceViewer]
	auctionData := &auction.AuctionData{}
	err = auctionData.RetrieveData(stub, issueEnd)
	stub.MockTransactionEnd("tx2")
	require.NoError(t, err, "Failed to retrieve auction data")

	var totalBuyBidQuantity int64 = 0
	for _, bid := range auctionData.BuyBids {
		totalBuyBidQuantity += bid.AskQuantity
	}

	// Run the auction
	auctionResult, err := auction.RunIndependent(auctionData)
	require.NoError(t, err, "Failed to run independent auction")

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

