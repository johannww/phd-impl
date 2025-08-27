package carbon_tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBid(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)
	stub.Creator = possibleIds[setup.REGULAR_ID]

	stub.TransientMap = map[string][]byte{
		"price": []byte("1000"),
	}

	stub.MockTransactionStart("tx1")
	err := bids.PublishBuyBidWithPublicQuanitity(stub, 100, &identities.X509Identity{})
	if err != nil {
		t.Fatalf("Error publishing buy bid: %v", err)
	}

	creatorId, _ := cid.GetID(stub)
	protoTs, _ := stub.GetTxTimestamp()
	lastInsertTimestamp := utils.TimestampRFC3339UtcString(protoTs)
	buyBid := &bids.BuyBid{}

	err = buyBid.FromWorldState(stub, []string{lastInsertTimestamp, creatorId})
	t.Log(buyBid.PrivatePrice)
	if err != nil || buyBid.PrivatePrice != nil {
		t.Fatalf(`PrivatePrice should be nil or error should not happen.
		REGULAR_ID should not be able to see it: %v`, err)
	}

	stub.Creator = possibleIds[identities.PriceViewer]
	creatorId, _ = cid.GetID(stub)
	buyBid = &bids.BuyBid{}

	err = buyBid.FromWorldState(stub, []string{lastInsertTimestamp, creatorId})
	if buyBid.PrivatePrice == nil {
		value, _, _ := cid.GetAttributeValue(stub, identities.PriceViewer)
		t.Logf("Error: %v", err)
		t.Logf("Creator's %s: %s", identities.PriceViewer, value)
		t.Fatal("PrivatePrice should be 1000. identities.PriceViewer should be able to see it")
	}
}

func TestBidBatchRecover(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)
	stub.Creator = possibleIds[setup.REGULAR_ID]

	numOfBids := int64(100)
	initialTime := time.Now()
	timeBeforeInsertion := utils.TimestampRFC3339UtcString(timestamppb.New(initialTime))
	for i := int64(0); i < numOfBids; i++ {
		stub.TransientMap = map[string][]byte{
			"price": []byte(strconv.FormatInt(i+10, 10)),
		}
		stub.MockTransactionStart("tx" + strconv.FormatInt(i, 10))
		stub.TxTimestamp = timestamppb.New(initialTime.Add(time.Duration(i) * time.Second))
		err := bids.PublishBuyBidWithPublicQuanitity(stub, 100, &identities.X509Identity{})
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

// func TestWithGoMock(t *testing.T) {
// 	ctrllr := gomock.NewController(t)
// 	mockStub := NewMockChaincodeStubInterface(ctrllr)
// 	t.Log(mockStub.GetTxID())
// 	_ = mockStub
// }
