package credits

import (
	"fmt"
	"math/rand"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const (
	LOCKED_CREDIT_PREFIX = "lockedCredit"
)

type LockedCredit struct {
	CreditID []string `json:"creditID"`
	Quantity int64    `json:"quantity"`
	LockID   string   `json:"lockID"`
}

var _ state.WorldStateManager = (*LockedCredit)(nil)

// FromWorldState implements state.WorldStateManager.
func (l *LockedCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	err := state.GetStateWithCompositeKey(stub, LOCKED_CREDIT_PREFIX, keyAttributes, l)
	if err != nil {
		return err
	}
	return nil
}

// GetID implements state.WorldStateManager.
func (l *LockedCredit) GetID() *[][]string {
	return &[][]string{l.CreditID}
}

// ToWorldState implements state.WorldStateManager.
func (l *LockedCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	if err := state.PutStateWithCompositeKey(stub, LOCKED_CREDIT_PREFIX, l.GetID(), l); err != nil {
		return fmt.Errorf("could not put locked credit in state: %v", err)
	}

	return nil

}

func LockCredit(stub shim.ChaincodeStubInterface, creditID []string, quantity int64) error {
	credit := &MintCredit{}
	err := credit.FromWorldState(stub, creditID)
	if err != nil {
		return fmt.Errorf("could not get credit from world state: %v", err)
	}

	if quantity == 0 {
		quantity = credit.Quantity
	}

	if quantity > credit.Quantity {
		return fmt.Errorf("cannot lock more credits than available: %d > %d", quantity, credit.Quantity)
	}

	// reduce the available quantity of the credit
	credit.Quantity -= quantity
	err = credit.ToWorldState(stub)
	if err != nil {
		return fmt.Errorf("could not update credit in world state: %v", err)
	}

	// TODOHP: think about the id later
	lockID := rand.Uint64()
	lockIDStr := fmt.Sprintf("%x", lockID)

	lockedCredit := &LockedCredit{
		CreditID: creditID,
		Quantity: quantity,
		LockID:   lockIDStr,
	}

	err = lockedCredit.ToWorldState(stub)
	if err != nil {
		return fmt.Errorf("could not put locked credit in state: %v", err)
	}

	return nil
}

func CreditIsLocked(stub shim.ChaincodeStubInterface, creditID []string) bool {
	lockedCredit := &LockedCredit{}
	err := lockedCredit.FromWorldState(stub, creditID)
	if err != nil {
		return false
	}

	return true
}
