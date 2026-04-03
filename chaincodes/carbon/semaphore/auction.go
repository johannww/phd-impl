// Package semaphore provides the shared auction-semaphore state used by both
// the auction and bids packages.  It is kept in its own package to avoid the
// import cycle that would arise if bids imported auction (which imports bids).
package semaphore

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
)

const (
	AUCTION_KEY = "auctionSemaphore"
)

// IsLocked reports whether the auction semaphore is currently set.
func IsLocked(stub shim.ChaincodeStubInterface) (bool, error) {
	val, err := stub.GetState(AUCTION_KEY)
	if err != nil {
		return false, fmt.Errorf("failed to read auction semaphore: %v", err)
	}
	return len(val) > 0, nil
}

// GetLockTimestamp returns the RFC3339 UTC timestamp stored in the semaphore,
// or an empty string if the auction is not currently locked.
func GetLockTimestamp(stub shim.ChaincodeStubInterface) (string, error) {
	val, err := stub.GetState(AUCTION_KEY)
	if err != nil {
		return "", fmt.Errorf("failed to read auction semaphore: %v", err)
	}
	return string(val), nil
}

// Lock writes the current transaction timestamp into the semaphore key.
// Callers are responsible for authorization checks before calling Lock.
func Lock(stub shim.ChaincodeStubInterface) error {
	txTS, err := stub.GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("could not get transaction timestamp: %v", err)
	}
	lockTS := utils.TimestampRFC3339UtcString(txTS)
	return stub.PutState(AUCTION_KEY, []byte(lockTS))
}

// Unlock removes the semaphore key from world state.
// Callers are responsible for authorization checks before calling Unlock.
func Unlock(stub shim.ChaincodeStubInterface) error {
	return stub.DelState(AUCTION_KEY)
}

// CheckAuctionAllowedTimestamp reports whether a bid is allowed given the
// current lock timestamp and the bid's own timestamp.
func CheckAuctionAllowedTimestamp(stub shim.ChaincodeStubInterface, bidTimestamp string) bool {
	lockTimestamp, _ := GetLockTimestamp(stub)
	if lockTimestamp == "" {
		return true
	}
	return bidTimestamp < lockTimestamp
}

// LockAuction sets the auction semaphore to the current transaction timestamp.
// Only identities with the TEEConfigurer attribute are authorized.
func LockAuction(stub shim.ChaincodeStubInterface) error {
	if err := cid.AssertAttributeValue(stub, identities.TEEConfigurer, "true"); err != nil {
		return fmt.Errorf("caller is not authorized to lock auction: %v", err)
	}
	return Lock(stub)
}

// UnlockAuction clears the auction semaphore.
// Only identities with the TEEConfigurer attribute are authorized.
func UnlockAuction(stub shim.ChaincodeStubInterface) error {
	if err := cid.AssertAttributeValue(stub, identities.TEEConfigurer, "true"); err != nil {
		return fmt.Errorf("caller is not authorized to unlock auction: %v", err)
	}
	return Unlock(stub)
}
