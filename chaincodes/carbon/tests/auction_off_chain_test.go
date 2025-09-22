package carbon_tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/stretchr/testify/require"
)

func TestOffChainIndependentAuction(t *testing.T) {
	stub, testData := genTestDataAndStub()

	issueStart, err := time.Parse(time.RFC3339, "2023-01-01T00:31:00Z")
	require.NoError(t, err)
	issueEnd := genAllMatchedBids(testData, issueStart, auction.AUCTION_INDEPENDENT)

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
	runner := &auction.AuctionIndepRunner{}
	auctionResultPub, auctionResultPvt, err := runner.RunIndependent(auctionData)
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
	runner := &auction.AuctionIndepRunner{}
	auctionResultPub, auctionResultPvt, err := runner.RunIndependent(auctionData)
	require.NoError(t, err, "Failed to run independent auction")
	auctionResult, err := auction.MergeIndependentPublicPrivateResults(auctionResultPub, auctionResultPvt)
	require.NoError(t, err, "Failed to merge independent auction results")

	verifyBidsQuantityConsistency(t, totalBuyBidQuantity, auctionResult)
}

func TestOffChainCoupledAuction(t *testing.T) {
	stub, testData := genTestDataAndStub()

	issueStart, err := time.Parse(time.RFC3339, "2023-01-01T00:31:00Z")
	require.NoError(t, err)
	issueEnd := genAllMatchedBids(testData, issueStart, auction.AUCTION_COUPLED)

	// Add policies for coupled auction
	testData.Policies = []policies.Name{policies.DISTANCE, policies.WIND_DIRECTION}

	stub.MockTransactionStart("tx1")
	stub.Creator = (*testData.Identities)[identities.PolicySetter]
	testData.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	setAuctionType(t, stub, auction.AUCTION_COUPLED, "tx2")

	auctionData := retriveAuctionDataFromWorldState(t, stub, issueEnd, testData, "tx3")

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
		bidAskQuantity := mintCredit.Quantity
		buyBid := &bids.BuyBid{
			BuyerID:     buyerIds[buyerIdIndex].Pseudonym,
			AskQuantity: bidAskQuantity,
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
