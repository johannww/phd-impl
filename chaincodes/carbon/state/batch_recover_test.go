package state

import (
	"strconv"
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetStatesPartialSecondaryIndex(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)

	stub.MockTransactionStart("tx1")
	numObjects := 10
	insertMockObjectsWithSecondaryIndex(stub, t, numObjects)
	stub.MockTransactionEnd("tx1")

	stub.MockTransactionStart("tx2")
	mockObjects, err := GetStateByPartialSecondaryIndex[MockObjectWithSecondaryIndex](stub, MOCK_OBJECT_PREFIX, []string{"secondary"})
	require.NoError(t, err, "Failed to get states by partial composite key")

	require.Len(t, mockObjects, numObjects, "Expected to retrieve all mock objects with secondary index")
}

func insertMockObjectsWithSecondaryIndex(stub *mocks.MockStub, t *testing.T, numObjects int) {
	for i := range numObjects {
		object := &MockObjectWithSecondaryIndex{
			MockAttr: strconv.Itoa(i),
			MockPvt:  "private_" + strconv.Itoa(i),
		}

		if err := object.ToWorldState(stub); err != nil {
			t.Fatalf("Failed to put state: %v", err)
		}
	}
}
