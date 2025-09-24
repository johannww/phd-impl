package credits

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

const (
	BURN_CREDIT_PREFIX = "BurnCredit"
)

// BurnCredit represents a minted carbon credit to be burned.
// it is associated to burn multiplier and burn timestamp.
type BurnCredit struct {
	MintCreditID  []string    `json:"mintCreditID"`
	MintCredit    *MintCredit `json:"mintCredit"`
	BurnQuantity  int64       `json:"burnQuantity"`
	BurnMult      int64       `json:"burnMult"`
	BurnTimeStamp string      `json:"burnTimestamp"`
}

func (bc *BurnCredit) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	// Load BurnCredit from world state
	if err := state.GetStateWithCompositeKey(stub, BURN_CREDIT_PREFIX, keyAttributes, bc); err != nil {
		return fmt.Errorf("failed to get BurnCredit from world state: %v", err)
	}

	// Load MintCredit from world state
	bc.MintCredit = &MintCredit{}
	if err := bc.MintCredit.FromWorldState(stub, bc.MintCreditID); err != nil {
		return fmt.Errorf("failed to load MintCredit data: %v", err)
	}

	return nil
}

func (bc *BurnCredit) ToWorldState(stub shim.ChaincodeStubInterface) error {
	copyBc := *bc // create a copy to avoid modifying the original object
	copyBc.MintCredit = nil
	if err := state.PutStateWithCompositeKey(stub, BURN_CREDIT_PREFIX, copyBc.GetID(), &copyBc); err != nil {
		return fmt.Errorf("could not put burn credit in state: %v", err)
	}

	// Ensure the associated MintCredit is also stored in the world state
	if err := bc.MintCredit.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not put associated MintCredit in state: %v", err)
	}

	return nil
}

func (bc *BurnCredit) GetID() *[][]string {
	return bc.MintCredit.GetID()
}

// TODO: implement
func Burn(stub shim.ChaincodeStubInterface) error {
	return nil
}
