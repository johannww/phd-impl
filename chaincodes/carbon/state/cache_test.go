package state

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
)

func TestGetStateFromCache(t *testing.T) {
	stub := mocks.NewMockStub("test", nil)

	stub.MockTransactionStart("tx1")

	// Create a mock object that implements the WorldStateManager interface
	mockObject := &worldStateObject{ID: [][]string{{"randomID", "2"}}}
	err := mockObject.ToWorldState(stub)
	if err != nil {
		t.Fatalf("Failed to put state: %v", err)
	}

	cache := map[string]*worldStateObject{}
	cachedValue, err := GetStateFromCache(stub, &cache, (*mockObject.GetID())[0])
	if err != nil {
		t.Fatalf("Failed to get state from cache: %v", err)
	}

	mustBeCachedValue, err := GetStateFromCache(stub, &cache, (*mockObject.GetID())[0])
	if err != nil {
		t.Fatalf("Failed to get state from cache: %v", err)
	}

	if mustBeCachedValue != cachedValue {
		t.Fatalf("Expected cached value to be the same as the one retrieved from world state")
	}

}

const MOCK_OBJECT_PREFIX = "mockObject"
const MOCK_OBJECT_PVT = "mockObjectPvt"

// worldStateObject is a base struct implementing WorldStateManager interface.
// It can be embedded in other structs to provide default implementations.
type worldStateObject struct {
	ID [][]string `json:"id"`
}

// GetID returns the ID of the object.
func (w *worldStateObject) GetID() *[][]string {
	return &w.ID
}

// FromWorldState loads the object from the world state using the keyAttributes.
func (w *worldStateObject) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := GetStateWithCompositeKey(stub, MOCK_OBJECT_PREFIX, keyAttributes, w)
	return err
}

// ToWorldState saves the object to the world state.
func (w *worldStateObject) ToWorldState(stub shim.ChaincodeStubInterface) error {
	err := PutStateWithCompositeKey(stub, MOCK_OBJECT_PREFIX, w.GetID(), w)
	return err

}

// worldStateObjectWithExtraPrefix is a base struct implementing WorldStateManagerWithExtraPrefix interface.
// It can be embedded in other structs to provide default implementations.
type worldStateObjectWithExtraPrefix struct {
	ID [][]string `json:"id"`
}

// GetID returns the ID of the object.
func (w *worldStateObjectWithExtraPrefix) GetID() *[][]string {
	return &[][]string{{"randonID", "2"}}
}

// FromWorldState loads the object from the world state using the keyAttributes and extraPrefix.
func (w *worldStateObjectWithExtraPrefix) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string, extraPrefix string) error {
	err := GetStateWithCompositeKey(stub, MOCK_OBJECT_PREFIX, keyAttributes, w)
	return err
}

// ToWorldState saves the object to the world state with extraPrefix.
func (w *worldStateObjectWithExtraPrefix) ToWorldState(stub shim.ChaincodeStubInterface, extraPrefix string) error {
	err := PutStateWithCompositeKey(stub, MOCK_OBJECT_PREFIX, w.GetID(), w)
	return err
}
