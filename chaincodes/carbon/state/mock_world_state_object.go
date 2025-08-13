package state

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

const (
	MOCK_OBJECT_PREFIX = "mockObject"
	MOCK_OBJECT_PVT    = "mockObjectPvt"
)

type MockObjectWithSecondaryIndex struct {
	MockAttr string `json:"mockAttr"`
	MockPvt  string `json:"mockPvtAttr"`
}

var _ WorldStateManager = (*MockObjectWithSecondaryIndex)(nil)

// FromWorldState implements WorldStateManager.
func (m *MockObjectWithSecondaryIndex) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	if err := GetStateWithCompositeKey(stub, MOCK_OBJECT_PREFIX, keyAttributes, m); err != nil {
		return err
	}

	if err := GetPvtDataWithCompositeKey(stub, MOCK_OBJECT_PREFIX, keyAttributes, MOCK_OBJECT_PVT, m); err != nil {
		return err
	}

	return nil
}

// GetID implements WorldStateManager.
func (m *MockObjectWithSecondaryIndex) GetID() *[][]string {
	return &[][]string{
		{m.MockAttr, "1"},
		{"secondary", m.MockAttr, "2"},
	}
}

// ToWorldState implements WorldStateManager.
func (m *MockObjectWithSecondaryIndex) ToWorldState(stub shim.ChaincodeStubInterface) error {
	copyMock := *m
	copyMock.MockPvt = "" // Temporarily unset MockPvt to avoid storing it in the public world state
	if err := PutStateWithCompositeKey(stub, MOCK_OBJECT_PREFIX, m.GetID(), &copyMock); err != nil {
		return err
	}

	if m.MockPvt == "" {
		return nil // No private data to store
	}

	if err := PutPvtDataWithCompositeKey(stub, MOCK_OBJECT_PREFIX, (*m.GetID())[0], MOCK_OBJECT_PVT, &m); err != nil {
		return err
	}

	return nil
}
