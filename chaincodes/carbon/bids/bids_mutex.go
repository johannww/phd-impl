package bids

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

// BidsMutex prevents write access to bids before a timestamp.
// This is useful for preventing phantom reads conflicts when
// getting all bids to post the commitment for the TEE.
type BidsMutex struct {
	Timestamp uint
}

const (
	BIDS_MUTEX_KEY = "bidsMutex"
)

func (b *BidsMutex) LockBidsBeforeTimestamp(stub shim.ChaincodeStubInterface) error {
	return state.PutState(stub, BIDS_MUTEX_KEY, b)
}
