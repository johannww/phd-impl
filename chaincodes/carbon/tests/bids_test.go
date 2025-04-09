package carbon_tests

import (
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/johannww/phd-impl/chaincodes/carbon"
	"github.com/johannww/phd-impl/chaincodes/carbon/bids"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
)

func TestBid(t *testing.T) {
	stub := shimtest.NewMockStub("carbon", &carbon.Carbon{})
	possibleIds := SetupIdentities(stub)
	stub.Creator = possibleIds[REGULAR_ID]
	stub.TxTimestamp = &timestamp.Timestamp{
		Seconds: 1,
	}

	stub.TransientMap = map[string][]byte{
		"price": []byte("1000"),
	}

	stub.MockTransactionStart("tx1")
	err := bids.PublishBuyBid(stub, 100, &identities.X509Identity{})
	if err != nil {
		t.Fatalf("Error publishing buy bid: %v", err)
	}
}

// func TestWithGoMock(t *testing.T) {
// 	ctrllr := gomock.NewController(t)
// 	mockStub := NewMockChaincodeStubInterface(ctrllr)
// 	t.Log(mockStub.GetTxID())
// 	_ = mockStub
// }
