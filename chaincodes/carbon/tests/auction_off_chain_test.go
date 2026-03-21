package carbon_tests

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	mocks "github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/stretchr/testify/require"
)

func TestOffChainIndependentAuction(t *testing.T) {
	nOwners := 10
	nChunks := 3
	nCompanies := 5
	startTimestamp := "2023-01-01T00:00:00Z"
	endTimestamp := "2023-01-01T00:30:00Z"
	issueInterval := 30 * time.Second
	stub, testData := genTestDataAndStub(
		nOwners, nChunks, nCompanies,
		startTimestamp, endTimestamp, issueInterval,
	)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	setAuctionType(t, stub, auction.AUCTION_INDEPENDENT, "tx2")

	auctionData := retriveAuctionDataFromWorldState(t, stub, testData, "tx3")

	var totalBuyBidQuantity int64 = 0
	for _, bid := range auctionData.BuyBids {
		totalBuyBidQuantity += bid.AskQuantity
	}

	// Run the auction
	runner := &auction.AuctionIndepRunner{}
	auctionResultPub, auctionResultPvt, err := runner.RunIndependent(auctionData)
	require.NoError(t, err, "Failed to run independent auction")
	auctionResult, err := auction.MergeIndependentPublicPrivateResults(auctionResultPub, auctionResultPvt)
	require.NoError(t, err, "Failed to merge independent auction results")

	verifyBidsQuantityConsistency(t, totalBuyBidQuantity, auctionResult)

	stub.MockTransactionStart("process-auction-result-tx")
	resultPubBytes, err := json.Marshal(auctionResultPub)
	require.NoError(t, err, "Failed to marshal public auction result")
	resultPvtBytes, err := json.Marshal(auctionResultPvt)
	require.NoError(t, err, "Failed to marshal private auction result")
	err = auction.ProcessOffChainAuctionResult(stub, resultPubBytes, resultPvtBytes)
	require.NoError(t, err, "Failed to process off-chain auction result")
	stub.MockTransactionEnd("process-auction-result-tx")
}

func TestOffChainIndependentAuctionWithRandomBids(t *testing.T) {
	nOwners := 10
	nChunks := 3
	nCompanies := 5
	startTimestamp := "2023-01-01T00:00:00Z"
	endTimestamp := "2023-01-01T00:30:00Z"
	issueInterval := 30 * time.Second
	stub, testData := genTestDataAndStub(
		nOwners, nChunks, nCompanies,
		startTimestamp, endTimestamp, issueInterval,
	)

	stub.MockTransactionStart("tx1")
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	setAuctionType(t, stub, auction.AUCTION_INDEPENDENT, "tx2")

	auctionData := retriveAuctionDataFromWorldState(t, stub, testData, "tx3")

	var totalBuyBidQuantity int64 = 0
	for _, bid := range auctionData.BuyBids {
		totalBuyBidQuantity += bid.AskQuantity
	}

	// Run the auction
	runner := &auction.AuctionIndepRunner{}
	auctionResultPub, auctionResultPvt, err := runner.RunIndependent(auctionData)
	require.NoError(t, err, "Failed to run independent auction")
	auctionResult, err := auction.MergeIndependentPublicPrivateResults(auctionResultPub, auctionResultPvt)
	require.NoError(t, err, "Failed to merge independent auction results")

	verifyBidsQuantityConsistency(t, totalBuyBidQuantity, auctionResult)
}

func TestOffChainCoupledAuction(t *testing.T) {
	nOwners := 10
	nChunks := 3
	nCompanies := 5
	startTimestamp := "2023-01-01T00:00:00Z"
	endTimestamp := "2023-01-01T00:30:00Z"
	issueInterval := 30 * time.Second
	stub, testData := genTestDataAndStub(
		nOwners, nChunks, nCompanies,
		startTimestamp, endTimestamp, issueInterval,
	)

	// Add policies for coupled auction
	testData.Policies = []policies.Name{policies.DISTANCE, policies.WIND_DIRECTION}

	stub.MockTransactionStart("tx1")
	stub.Creator = (*testData.Identities)[identities.PolicySetter]
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	setAuctionType(t, stub, auction.AUCTION_COUPLED, "tx2")

	auctionData := retriveAuctionDataFromWorldState(t, stub, testData, "tx3")

	var totalBuyBidQuantity int64 = 0
	for _, bid := range auctionData.BuyBids {
		totalBuyBidQuantity += bid.PrivateQuantity.AskQuantity
	}

	pApplier := &policies.PolicyApplierImpl{
		DefinedPolicies: make(map[policies.Name]policies.PolicyFunc),
	}
	// Run the auction
	thirtyPercentPolicy := func(input *policies.PolicyInput) int64 {
		return 300 // 30% increase
	}
	pApplier.DefinedPolicies[policies.DISTANCE] = thirtyPercentPolicy
	pApplier.DefinedPolicies[policies.WIND_DIRECTION] = thirtyPercentPolicy
	runner := &auction.AuctionCoupledRunner{}
	auctionResultPub, auctionResultPvt, err := runner.RunCoupled(auctionData, pApplier)
	require.NoError(t, err, "Failed to run coupled auction")

	auctionResult, err := auction.NewSingleCoupledResults(auctionResultPub, auctionResultPvt)
	require.NoError(t, err, "Failed to merge coupled auction results")
	mergedMatchedBids, err := auctionResult.MergeIntoSingleMatchedBids()
	require.NoError(t, err, "Failed to merge matched bids into single list containing the public and private parts")
	verifyMultiplierMultiplyAsExpected(t, mergedMatchedBids,
		[]policies.PolicyFunc{thirtyPercentPolicy, thirtyPercentPolicy})
	verifyPrivateDataIsInPrivatePart(t, auctionResultPub, auctionResultPvt)

	adjustedSellBids, adjustedBuyBids := auctionResult.MergeIntoSingleAdjustedBids()
	verifyAdjustedBidsConsistency(t, mergedMatchedBids, adjustedSellBids, adjustedBuyBids)

	stub.MockTransactionStart("process-auction-result-tx")
	resultPubBytes, err := json.Marshal(auctionResultPub)
	require.NoError(t, err, "Failed to marshal public auction result")
	resultPvtBytes, err := json.Marshal(auctionResultPvt)
	require.NoError(t, err, "Failed to marshal private auction result")
	err = auction.ProcessOffChainAuctionResult(stub, resultPubBytes, resultPvtBytes)
	require.NoError(t, err, "Failed to process off-chain auction result")
	stub.MockTransactionEnd("process-auction-result-tx")

	// verify matched bids exist in world state after processing auction result
	stub.MockTransactionStart("get-matched-bids-tx")
	matchedBids, err := state.GetStatesByPartialCompositeKey[bids.MatchedBid](stub, bids.MATCHED_BID_PREFIX, []string{})
	require.NoError(t, err, "Failed to get matched bids from world state")
	require.NotZero(t, len(matchedBids), "There should be matched bids in world state after processing auction result")
	require.Equal(t, len(mergedMatchedBids), len(matchedBids), "Number of matched bids in world state should match number of merged matched bids")
	stub.MockTransactionEnd("get-matched-bids-tx")
}

// verifyMultiplierMultiplyAsExpected tests the multiplying logic.
// It uses float64 just to double-check our fixed point int64 arithmetic
func verifyMultiplierMultiplyAsExpected(t *testing.T,
	mergedMatchedBids []*bids.MatchedBid,
	policiesApplied []policies.PolicyFunc) {

	for i, mb := range mergedMatchedBids {
		floatMultiplier := 1.0
		for _, policyFunc := range policiesApplied {
			floatMultiplier *= 1.0 + (float64(policyFunc(nil)) / 1000.0)
		}

		// require.InDelta(t, floatMultiplier, 1.0+float64(matchedBid.PrivateMultiplier.Value)/float64(matchedBid.PrivateMultiplier.Scale), 0.001, "In matching %d, multiplier %.3f should be close to private multiplier %d divided by 1000", i, floatMultiplier, matchedBid.PrivateMultiplier)

		require.LessOrEqual(t, mb.Quantity,
			mb.BuyBid.PrivateQuantity.AskQuantity,
			"In matching %d, matched quantity %d should not be greater than"+
				" asked quantity %d", i, mb.Quantity,
			mb.BuyBid.PrivateQuantity.AskQuantity)

		minQuantity := min(mb.BuyBid.PrivateQuantity.AskQuantity, mb.SellBid.Quantity)
		acquirableQuantity := floatMultiplier * float64(minQuantity)
		expectedNominalQuantity := 0.0
		if mb.BuyBid.PrivateQuantity.AskQuantity >= int64(acquirableQuantity) {
			expectedNominalQuantity = float64(minQuantity)
		} else {
			expectedNominalQuantity = float64(minQuantity) / floatMultiplier
		}

		require.InDelta(t, float64(mb.Quantity), expectedNominalQuantity, 2.0,
			"In matching %v, with minQuantity %.3f matched quantity %d should be close to expected "+
				"nominal quantity %.2f", mb, float64(minQuantity), mb.Quantity, expectedNominalQuantity)
	}

}

func verifyPrivateDataIsInPrivatePart(t *testing.T, auctionResultPub, auctionResultPvt *auction.OffChainCoupledAuctionResult) {
	for i, pubMatchedBid := range auctionResultPub.MatchedBidsPublic {
		privMatchedBid := auctionResultPvt.MatchedBidsPrivate[i]

		require.Nil(t, pubMatchedBid.BuyBid.PrivatePrice, "Public part of matched bid %d should not have private price for buy bid", i)
		require.Nil(t, pubMatchedBid.BuyBid.PrivateQuantity, "Public part of matched bid %d should not have private quantity for buy bid", i)
		require.Nil(t, pubMatchedBid.SellBid.PrivatePrice, "Public part of matched bid %d should not have private price for sell bid", i)

		require.NotNil(t, privMatchedBid.BuyBid.PrivatePrice, "Private part of matched bid %d should have private price for buy bid", i)
		require.NotNil(t, privMatchedBid.BuyBid.PrivateQuantity, "Private part of matched bid %d should have private quantity for buy bid", i)
		require.NotNil(t, privMatchedBid.SellBid.PrivatePrice, "Private part of matched bid %d should have private price for sell bid", i)

		require.Nil(t, pubMatchedBid.PrivateMultiplier, "Public part of matched bid %d should not have private multiplier", i)
		require.Nil(t, pubMatchedBid.PrivatePrice, "Public part of matched bid %d should not have private price", i)
		require.NotNil(t, privMatchedBid.PrivateMultiplier, "Private part of matched bid %d should have private multiplier", i)
		require.NotNil(t, privMatchedBid.PrivatePrice, "Private part of matched bid %d should have private price", i)
	}

	for i, pubBuyBid := range auctionResultPub.AdjustedBuyBidsPublic {
		privBuyBid := auctionResultPvt.AdjustedBuyBidsPrivate[i]
		require.Nil(t, pubBuyBid.PrivatePrice, "Public part of adjusted buy bid %d should not have private price", i)
		require.Nil(t, pubBuyBid.PrivateQuantity, "Public part of adjusted buy bid %d should not have private quantity", i)
		require.NotNil(t, privBuyBid.PrivatePrice, "Private part of adjusted buy bid %d should have private price", i)
		require.NotNil(t, privBuyBid.PrivateQuantity, "Private part of adjusted buy bid %d should have private quantity", i)
	}

	for i, pubSellBid := range auctionResultPub.AdjustedSellBidsPublic {
		privSellBid := auctionResultPvt.AdjustedSellBidsPrivate[i]
		require.Nil(t, pubSellBid.PrivatePrice, "Public part of adjusted sell bid %d should not have private price", i)
		require.NotNil(t, privSellBid.PrivatePrice, "Private part of adjusted sell bid %d should have private price", i)
	}
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

func genTestDataAndStub(
	nOwners int, nChunks int, nCompanies int,
	startTimestamp string, endTimestamp string,
	issueInterval time.Duration,
) (*mocks.MockStub, *utils_test.TestData) {
	testData := utils_test.GenDataWithBids(
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
	t *testing.T, stub *mocks.MockStub,
	testData *utils_test.TestData,
	txID string,
) *auction.AuctionData {
	stub.MockTransactionStart(txID)
	stub.Creator = (*testData.Identities)[identities.PriceViewer]
	auctionData := &auction.AuctionData{}
	auctionID, err := auction.IncrementAuctionID(stub)
	require.NoError(t, err, "Failed to increment auction ID")
	err = auctionData.RetrieveData(stub, testData.BidIssueLastTs)
	auctionData.AuctionID = auctionID
	stub.MockTransactionEnd(txID)
	require.NoError(t, err, "Failed to retrieve auction data")
	return auctionData
}

func verifyAdjustedBidsConsistency(
	t *testing.T,
	mergedMatchedBids []*bids.MatchedBid,
	adjSellBids []*bids.SellBid,
	adjBuyBids []*bids.BuyBid,
) {
	sellMatchedTotal := make(map[string]int64)
	sellInitialQuantity := make(map[string]int64)
	buyMatchedTotal := make(map[string]int64)
	buyInitialQuantity := make(map[string]int64)

	for _, mb := range mergedMatchedBids {
		require.NotNil(t, mb.SellBid)
		require.NotNil(t, mb.BuyBid)
		require.NotNil(t, mb.BuyBid.PrivateQuantity)

		sellID := strings.Join((*mb.SellBid.GetID())[0], "|")
		buyID := strings.Join((*mb.BuyBid.GetID())[0], "|")

		sellMatchedTotal[sellID] += mb.Quantity
		buyMatchedTotal[buyID] += mb.Quantity

		// Get the maximum initial quantity, as this represents the initial quantity before any matches, which is what the adjusted quantity should be based on
		if mb.SellBid.Quantity > sellInitialQuantity[sellID] {
			sellInitialQuantity[sellID] = mb.SellBid.Quantity
		}
		if mb.BuyBid.PrivateQuantity.AskQuantity > buyInitialQuantity[buyID] {
			buyInitialQuantity[buyID] = mb.BuyBid.PrivateQuantity.AskQuantity
		}
	}

	seenAdjustedSell := make(map[string]bool)
	for _, sb := range adjSellBids {
		sellID := strings.Join((*sb.GetID())[0], "|")
		seenAdjustedSell[sellID] = true
		expectedAdjusted := sellInitialQuantity[sellID] - sellMatchedTotal[sellID]
		require.Equalf(t, expectedAdjusted, sb.Quantity,
			"adjusted sell quantity for %s should match initial-minus-matched", sellID)
	}

	seenAdjustedBuy := make(map[string]bool)
	for _, bb := range adjBuyBids {
		require.NotNil(t, bb.PrivateQuantity)
		buyID := strings.Join((*bb.GetID())[0], "|")
		seenAdjustedBuy[buyID] = true
		expectedAdjusted := buyInitialQuantity[buyID] - buyMatchedTotal[buyID]
		require.Equalf(t, expectedAdjusted, bb.PrivateQuantity.AskQuantity,
			"adjusted buy quantity for %s should match initial-minus-matched", buyID)
	}

	require.Equal(t, len(sellMatchedTotal), len(seenAdjustedSell),
		"adjusted sell bids should contain exactly the bids that were matched")
	require.Equal(t, len(buyMatchedTotal), len(seenAdjustedBuy),
		"adjusted buy bids should contain exactly the bids that were matched")
}
