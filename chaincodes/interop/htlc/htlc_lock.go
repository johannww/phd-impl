package htlc

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const HTLC_LOCK_PREFIX = "htlcLock"

// HtlcLock represents a lock credit is designated for in an HTLC
type HtlcLock struct {
	LockID string `json:"lockID"`
	HTLCID string `json:"htlcID"`
}

var _ state.WorldStateManager = (*HtlcLock)(nil)

// FromWorldState implements state.WorldStateManager.
func (h *HtlcLock) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, HTLC_LOCK_PREFIX, keyAttributes, h)
}

// ToWorldState implements state.WorldStateManager.
func (h *HtlcLock) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, HTLC_LOCK_PREFIX, h.GetID(), h)
}

func (h *HtlcLock) DeleteFromWorldState(stub shim.ChaincodeStubInterface) error {
	return state.DeleteStateWithCompositeKey(stub, HTLC_LOCK_PREFIX, h.GetID())
}

// GetID implements state.WorldStateManager.
func (h *HtlcLock) GetID() *[][]string {
	return &[][]string{{h.LockID, h.HTLCID}}
}
