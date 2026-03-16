package carbon_tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/chaincodes/carbon/payment"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	utils_test "github.com/johannww/phd-impl/chaincodes/carbon/tests/utils"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	mocks "github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBidWithNoWallet(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)
	stub.Creator = possibleIds[setup.REGULAR_ID]

	stub.TransientMap = map[string][]byte{
		"price": []byte("1000"),
	}

	stub.MockTransactionStart("tx1")
	err := bids.PublishBuyBidWithPublicQuanitity(stub, 100)
	require.Error(t, err, "could not get buyer wallet")
}

func TestBid(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)
	stub.Creator = possibleIds[setup.IDEMIX_ID]

	stub.TransientMap = map[string][]byte{
		"price": []byte("1000"),
	}

	stub.MockTransactionStart("tx1")
	createAndStoreWalletForCreator(t, stub, 200000000)

	err := bids.PublishBuyBidWithPublicQuanitity(stub, 100)
	require.NoError(t, err, "Error publishing buy bid")

	creatorId := identities.GetID(stub)
	protoTs, _ := stub.GetTxTimestamp()
	lastInsertTimestamp := utils.TimestampRFC3339UtcString(protoTs)
	buyBid := &bids.BuyBid{}

	err = buyBid.FromWorldState(stub, []string{lastInsertTimestamp, creatorId})
	t.Log(buyBid.PrivatePrice)
	require.NoError(t, err)
	require.NotNil(t, buyBid.PrivatePrice, "Bid owner should be able to see private price")

	stub.Creator = possibleIds[setup.REGULAR_ID]
	buyBid = &bids.BuyBid{} // reset buyBid
	err = buyBid.FromWorldState(stub, []string{lastInsertTimestamp, creatorId})
	require.NoError(t, err)
	require.Nil(t, buyBid.PrivatePrice, "Non bid owner should not be able to see private price")

	stub.Creator = possibleIds[identities.PriceViewer]
	buyBid = &bids.BuyBid{}

	err = buyBid.FromWorldState(stub, []string{lastInsertTimestamp, creatorId})
	if buyBid.PrivatePrice == nil {
		value, _, _ := cid.GetAttributeValue(stub, identities.PriceViewer)
		t.Logf("Error: %v", err)
		t.Logf("Creator's %s: %s", identities.PriceViewer, value)
		t.Fatal("PrivatePrice should be 1000. identities.PriceViewer should be able to see it")
	}

	// Also validate private-quantity flow in the same test.
	stub.MockTransactionEnd("tx1")
	stub.MockTransactionStart("tx2")
	stub.TxTimestamp = timestamppb.New(protoTs.AsTime().Add(1 * time.Second))
	stub.Creator = possibleIds[setup.IDEMIX_ID]
	stub.TransientMap = map[string][]byte{
		"price":    []byte("1100"),
		"quantity": []byte("75"),
	}

	err = bids.PublishBuyBidWithPrivateQuantity(stub)
	require.NoError(t, err, "Error publishing buy bid with private quantity")

	privateProtoTs, _ := stub.GetTxTimestamp()
	privateInsertTimestamp := utils.TimestampRFC3339UtcString(privateProtoTs)

	// Owner should see private price and private quantity, while public quantity remains hidden.
	privateQtyBid := &bids.BuyBid{}
	err = privateQtyBid.FromWorldState(stub, []string{privateInsertTimestamp, creatorId})
	require.NoError(t, err)
	require.Equal(t, int64(0), privateQtyBid.AskQuantity)
	require.NotNil(t, privateQtyBid.PrivatePrice)
	require.NotNil(t, privateQtyBid.PrivateQuantity)
	require.Equal(t, int64(75), privateQtyBid.PrivateQuantity.AskQuantity)

	// Non-owner/non-PriceViewer should not see private fields.
	stub.Creator = possibleIds[setup.REGULAR_ID]
	otherViewPrivateQtyBid := &bids.BuyBid{}
	err = otherViewPrivateQtyBid.FromWorldState(stub, []string{privateInsertTimestamp, creatorId})
	require.NoError(t, err)
	require.Nil(t, otherViewPrivateQtyBid.PrivatePrice)
	require.Nil(t, otherViewPrivateQtyBid.PrivateQuantity)

	// PriceViewer should see private fields.
	stub.Creator = possibleIds[identities.PriceViewer]
	priceViewerPrivateQtyBid := &bids.BuyBid{}
	err = priceViewerPrivateQtyBid.FromWorldState(stub, []string{privateInsertTimestamp, creatorId})
	require.NoError(t, err)
	require.NotNil(t, priceViewerPrivateQtyBid.PrivatePrice)
	require.NotNil(t, priceViewerPrivateQtyBid.PrivateQuantity)
	require.Equal(t, int64(75), priceViewerPrivateQtyBid.PrivateQuantity.AskQuantity)
}

func TestSellBidOwnerCanReadPrivatePrice(t *testing.T) {
	nOwners := 1
	nChunks := 3
	nCompanies := 0
	startTimestamp := "2023-01-01T00:00:00Z"
	endTimestamp := "2023-01-01T00:05:00Z"
	issueInterval := 30 * time.Second
	stub, testdata := genTestDataAndStub(
		nOwners, nChunks, nCompanies,
		startTimestamp, endTimestamp, issueInterval,
	)
	stub.MockTransactionStart("tx1")
	testdata.SaveToWorldState(stub)
	stub.MockTransactionEnd("tx1")

	possibleIds := *testdata.Identities
	ownerID := possibleIds[utils_test.OWNER_PREFIX+"0"]
	stub.Creator = ownerID
	creditID := (*testdata.MintCredits[0].GetID())[0]
	initialCreditQuantity := testdata.MintCredits[0].Quantity
	sellerID := identities.GetID(stub)

	sellQuantity := int64(4000)
	stub.TransientMap = map[string][]byte{
		"price": []byte("1234"),
	}

	stub.MockTransactionStart("tx-sell-1")
	err := bids.PublishSellBid(stub, sellQuantity, creditID)
	require.NoError(t, err)
	protoTs, err := stub.GetTxTimestamp()
	require.NoError(t, err)
	bidTS := utils.TimestampRFC3339UtcString(protoTs)
	stub.MockTransactionEnd("tx-sell-1")

	bidKey := []string{bidTS, sellerID}

	// Owner can read private price.
	stub.MockTransactionStart("tx-sell-owner-read")
	stub.Creator = ownerID
	ownerView := &bids.SellBid{}
	err = ownerView.FromWorldState(stub, bidKey)
	require.NoError(t, err)
	require.NotNil(t, ownerView.PrivatePrice, "seller owner should be able to read private sell price")
	require.Equal(t, int64(1234), ownerView.PrivatePrice.Price)
	stub.MockTransactionEnd("tx-sell-owner-read")

	// Non-owner/non-PriceViewer cannot read private price.
	stub.MockTransactionStart("tx-sell-other-read")
	stub.Creator = possibleIds[setup.IDEMIX_ID]
	otherView := &bids.SellBid{}
	err = otherView.FromWorldState(stub, bidKey)
	require.NoError(t, err)
	require.Nil(t, otherView.PrivatePrice, "non-owner without PriceViewer should not read private sell price")
	stub.MockTransactionEnd("tx-sell-other-read")

	// PriceViewer can still read private price.
	stub.MockTransactionStart("tx-sell-priceviewer-read")
	stub.Creator = possibleIds[identities.PriceViewer]
	adminView := &bids.SellBid{}
	err = adminView.FromWorldState(stub, bidKey)
	require.NoError(t, err)
	require.NotNil(t, adminView.PrivatePrice, "PriceViewer should read private sell price")
	require.Equal(t, int64(1234), adminView.PrivatePrice.Price)
	stub.MockTransactionEnd("tx-sell-priceviewer-read")

	// Non-owner cannot retract the sell bid.
	stub.MockTransactionStart("tx-sell-other-retract")
	stub.Creator = possibleIds[setup.IDEMIX_ID]
	err = bids.RetractSellBid(stub, bidKey)
	require.Error(t, err, "non-owner should not be able to retract sell bid")
	stub.MockTransactionEnd("tx-sell-other-retract")

	// Owner retracts sell bid, which restores credit quantity and deletes the bid.
	stub.MockTransactionStart("tx-sell-owner-retract")
	stub.Creator = ownerID
	err = bids.RetractSellBid(stub, bidKey)
	require.NoError(t, err)

	credit := &credits.MintCredit{}
	err = credit.FromWorldState(stub, creditID)
	require.NoError(t, err)
	require.Equal(t, initialCreditQuantity, credit.Quantity, "credit quantity should be restored after sell bid retract")
	stub.MockTransactionEnd("tx-sell-owner-retract")

	stub.MockTransactionStart("tx-sell-check-deleted")
	deletedBid := &bids.SellBid{}
	err = deletedBid.FromWorldState(stub, bidKey)
	require.Error(t, err, "sell bid should be deleted after retract")
	stub.MockTransactionEnd("tx-sell-check-deleted")
}

func TestRetractBidRefundsWalletAndDeletesBid(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)
	stub.Creator = possibleIds[setup.IDEMIX_ID]

	initialWallet := int64(200000000)
	price := int64(1000)
	quantity := int64(100)

	stub.MockTransactionStart("tx1")
	createAndStoreWalletForCreator(t, stub, initialWallet)
	stub.TransientMap = map[string][]byte{
		"price": []byte(strconv.FormatInt(price, 10)),
	}
	err := bids.PublishBuyBidWithPublicQuanitity(stub, quantity)
	require.NoError(t, err, "Error publishing buy bid")

	creatorID := identities.GetID(stub)
	protoTs, err := stub.GetTxTimestamp()
	require.NoError(t, err)
	bidTS := utils.TimestampRFC3339UtcString(protoTs)
	bidID := []string{bidTS, creatorID}

	// Wallet should be debited after placing the bid.
	wallet := &payment.VirtualTokenWallet{}
	err = wallet.FromWorldState(stub, []string{creatorID})
	require.NoError(t, err)
	require.Equal(t, initialWallet-price*quantity, wallet.Quantity)

	err = bids.RetractBuyBid(stub, bidID)
	require.NoError(t, err, "Error retracting buy bid")
	stub.MockTransactionEnd("tx1")

	// Wallet should be restored after retract.
	stub.MockTransactionStart("tx2")
	err = wallet.FromWorldState(stub, []string{creatorID})
	require.NoError(t, err)
	require.Equal(t, initialWallet, wallet.Quantity)

	// Bid should be deleted.
	buyBid := &bids.BuyBid{}
	err = buyBid.FromWorldState(stub, bidID)
	require.Error(t, err, "buy bid should be deleted after retract")
	stub.MockTransactionEnd("tx2")
}

func TestBidBatchRecover(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)
	stub.Creator = possibleIds[setup.REGULAR_ID]
	createAndStoreWalletForCreator(t, stub, 200000000)

	numOfBids := int64(100)
	initialTime := time.Now()
	timeBeforeInsertion := utils.TimestampRFC3339UtcString(timestamppb.New(initialTime))
	for i := int64(0); i < numOfBids; i++ {
		stub.TransientMap = map[string][]byte{
			"price": []byte(strconv.FormatInt(i+10, 10)),
		}
		stub.MockTransactionStart("tx" + strconv.FormatInt(i, 10))
		stub.TxTimestamp = timestamppb.New(initialTime.Add(time.Duration(i) * time.Second))
		err := bids.PublishBuyBidWithPublicQuanitity(stub, 100)
		if err != nil {
			t.Fatalf("Error publishing buy bid: %v", err)
		}
	}
	timeAfterInsertion := utils.TimestampRFC3339UtcString(timestamppb.New(stub.TxTimestamp.AsTime().Add(time.Duration(1) * time.Second)))

	buyBids, err := state.GetStatesByRangeCompositeKey[bids.BuyBid](stub, bids.BUY_BID_PREFIX, []string{timeBeforeInsertion}, []string{timeAfterInsertion})
	if err != nil {
		t.Fatalf("Error getting buy bids by range: %v", err)
	}

	ensureAllBidsWereRetrieved(stub, t, buyBids, int(numOfBids), possibleIds)

	buyBids, err = state.GetStatesByPartialCompositeKey[bids.BuyBid](stub, bids.BUY_BID_PREFIX, nil)
	if err != nil {
		t.Fatalf("Error getting buy bids by range: %v", err)
	}
	ensureAllBidsWereRetrieved(stub, t, buyBids, int(numOfBids), possibleIds)
}

func ensureAllBidsWereRetrieved(stub *mocks.MockStub, t *testing.T,
	buyBids []*bids.BuyBid, numOfBids int,
	possibleIds setup.MockIdentities) {
	if len(buyBids) != int(numOfBids) {
		t.Fatalf("Expected %d buy bids, got %d", numOfBids, len(buyBids))
	}

	// recover private prices
	for i, bid := range buyBids {
		i := int64(i)
		stub.Creator = possibleIds[identities.PriceViewer]
		bid.FetchPrivatePrice(stub)
		if bid.PrivatePrice == nil {
			t.Fatalf("PriceViewer should be able to see the private price: %v", bid.PrivatePrice)
		}
		// t.Logf("BuyBid %d: %v\n", i, bid)
		// t.Logf("PrivatePrice %d: %v\n", i, bid.PrivatePrice)
		if bid.PrivatePrice.Price != i+10 {
			t.Fatalf("Expected private price %d, got %d", i+10, bid.PrivatePrice.Price)
		}
	}
}

func createAndStoreWalletForCreator(t *testing.T, stub *mocks.MockStub, currencyQuantity int64) {
	ownerID := identities.GetID(stub)
	buyerWallet := &payment.VirtualTokenWallet{
		OwnerID:  ownerID,
		Quantity: currencyQuantity,
	}
	err := buyerWallet.ToWorldState(stub)
	require.NoError(t, err, "Failed to store buyer wallet in world state")
}
