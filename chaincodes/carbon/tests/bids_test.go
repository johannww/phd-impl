package carbon_tests

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/johannww/phd-impl/chaincodes/carbon"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
)

func TestBid(t *testing.T) {
	stub := shimtest.NewMockStub("carbon", &carbon.Carbon{})
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
	lastInsertTimestamp := stub.TxTimestamp.AsTime().String()
	buyBid := &bids.BuyBid{}

	err = buyBid.FromWorldState(stub, []string{creatorId, lastInsertTimestamp})
	t.Log(buyBid.PrivatePrice)
	if err != nil || buyBid.PrivatePrice != nil {
		t.Fatal("PrivatePrice should be nil. REGULAR_ID should not be able to see it")
	}

	stub.Creator = possibleIds[identities.PriceViewer]
	creatorId, _ = cid.GetID(stub)
	buyBid = &bids.BuyBid{}

	err = buyBid.FromWorldState(stub, []string{creatorId, lastInsertTimestamp})
	if buyBid.PrivatePrice == nil {
		value, _, _ := cid.GetAttributeValue(stub, identities.PriceViewer)
		t.Logf("Error: %v", err)
		t.Logf("Creator's %s: %s", identities.PriceViewer, value)
		t.Fatal("PrivatePrice should be 1000. identities.PriceViewer should be able to see it")
	}
}

// func TestWithGoMock(t *testing.T) {
// 	ctrllr := gomock.NewController(t)
// 	mockStub := NewMockChaincodeStubInterface(ctrllr)
// 	t.Log(mockStub.GetTxID())
// 	_ = mockStub
// }
