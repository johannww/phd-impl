package carbon_tests

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/johannww/phd-impl/chaincodes/carbon"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBid(t *testing.T) {
	stub := NewMockStub("carbon", &carbon.Carbon{})
	possibleIds := SetupIdentities(stub)
	stub.Creator = possibleIds[REGULAR_ID]

	stub.TransientMap = map[string][]byte{
		"price": []byte("1000"),
	}

	stub.MockTransactionStart("tx1")
	err := bids.PublishBuyBid(stub, 100, &identities.X509Identity{})
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
	stub := NewMockStub("carbon", &carbon.Carbon{})
	possibleIds := SetupIdentities(stub)
	stub.Creator = possibleIds[REGULAR_ID]

	numOfBids := int64(100)
	timeBeforeInsertion := utils.TimestampRFC3339UtcString(timestamppb.Now())
	for i := int64(0); i < numOfBids; i++ {
		stub.TransientMap = map[string][]byte{
			"price": []byte(strconv.FormatInt(i+10, 10)),
		}
		stub.MockTransactionStart("tx" + strconv.FormatInt(i, 10))
		stub.TxTimestamp = timestamppb.New(time.Now().Add(time.Duration(i) * time.Second))
		err := bids.PublishBuyBid(stub, 100, &identities.X509Identity{})
		if err != nil {
			t.Fatalf("Error publishing buy bid: %v", err)
		}
	}
	oneSecDurtaion, _ := time.ParseDuration("1s")
	timeAfterInsertion := utils.TimestampRFC3339UtcString(timestamppb.New(stub.TxTimestamp.AsTime().Add(oneSecDurtaion)))

	buyBidsBytes, err := state.GetStatesByRangeCompositeKey(stub, bids.BUY_BID_PREFIX, []string{timeBeforeInsertion}, []string{timeAfterInsertion})
	if err != nil {
		t.Fatalf("Error getting buy bids by range: %v", err)
	}

	buyBids := make([]bids.BuyBid, len(buyBidsBytes))
	for i, bid := range buyBidsBytes {
		err = json.Unmarshal(bid, &buyBids[i])
		if err != nil {
			t.Fatalf("Error unmarshalling buy bid: %v", err)
		}

		buyBidID := (*buyBids[i].GetID())[0]
		if buyBids[i].FromWorldState(stub, buyBidID) != nil {
			t.Fatalf("BuyBid from range query is not the same as the existing in the world state: %v", err)
		}

	}
}

// func TestWithGoMock(t *testing.T) {
// 	ctrllr := gomock.NewController(t)
// 	mockStub := NewMockChaincodeStubInterface(ctrllr)
// 	t.Log(mockStub.GetTxID())
// 	_ = mockStub
// }
