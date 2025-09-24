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

// TODO: Test
// BurnQuantity handles the burning of a minted credit.
func BurnQuantity(stub shim.ChaincodeStubInterface, mintCreditID []string, burnQuantity int64) error {
	mc, err := loadMintCreditAndPerformChecks(stub, burnQuantity, mintCreditID)
	if err != nil {
		return err
	}

	bc := &BurnCredit{
		BurnQuantity: burnQuantity,
		MintCreditID: mintCreditID,
		MintCredit:   mc,
	}

	if err = fillTSAndCalcMult(stub, bc); err != nil {
		return fmt.Errorf("could not fill timestamp and calculate multiplier: %v", err)
	}

	bc.MintCredit.Credit.Quantity -= burnQuantity
	if err = bc.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not update mint credit in world state: %v", err)
	}

	if err = bc.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not put burn credit in world state: %v", err)
	}

	return nil
}

func loadMintCreditAndPerformChecks(
	stub shim.ChaincodeStubInterface,
	burnQuantity int64,
	mintCreditID []string) (*MintCredit, error) {
	mc := &MintCredit{}
	err := mc.FromWorldState(stub, mintCreditID)
	if err != nil {
		return nil, fmt.Errorf("could not get mint credit from world state: %v", err)
	}

	if burnQuantity > mc.Quantity {
		return nil, fmt.Errorf("burn quantity exceeds available quantity: %d > %d", burnQuantity, mc.Quantity)
	}

	callerID := identities.GetID(stub)
	if mc.OwnerID != callerID {
		return nil, fmt.Errorf("only the owner of the credit can burn it: %s != %s", mc.OwnerID, callerID)
	}
	return mc, nil
}

func fillTSAndCalcMult(stub shim.ChaincodeStubInterface, bc *BurnCredit) error {
	protoTs, _ := stub.GetTxTimestamp()
	burnTimestamp := utils.TimestampRFC3339UtcString(protoTs)
	bc.BurnTimeStamp = burnTimestamp

	pApplier := policies.NewPolicyApplier()
	// TODOHP: the multiplier here might expose buyer real identity.
	// Evaluate this later.
	pInput := &policies.PolicyInput{}
	activePols, err := policies.GetActivePolicies(stub)
	if err != nil {
		return fmt.Errorf("could not get active policies: %v", err)
	}
	bc.BurnMult, err = pApplier.BurnIndependentMult(pInput, activePols)
	if err != nil {
		return fmt.Errorf("could not get burn multiplier: %v", err)
	}

	return nil

}
