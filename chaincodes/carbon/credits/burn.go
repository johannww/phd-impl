package credits

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/policies"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
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
	Adjusted      bool        `json:"adjusted"`
}

var _ state.WorldStateManager = (*BurnCredit)(nil)

func (bc *BurnCredit) FromProto(m proto.Message) error {
	pbBc, ok := m.(*pb.BurnCredit)
	if !ok {
		return fmt.Errorf("unexpected proto message type for BurnCredit")
	}

	bc.MintCreditID = pbBc.MintCreditID
	bc.BurnQuantity = pbBc.BurnQuantity
	bc.BurnMult = pbBc.BurnMult
	bc.BurnTimeStamp = pbBc.BurnTimestamp
	bc.Adjusted = pbBc.Adjusted

	// Convert MintCredit
	if pbBc.MintCredit != nil {
		mc := &MintCredit{}
		if err := mc.FromProto(pbBc.MintCredit); err != nil {
			return fmt.Errorf("could not convert MintCredit from proto: %v", err)
		}
		bc.MintCredit = mc
	}

	return nil
}

func (bc *BurnCredit) ToProto() proto.Message {
	var pbMintCredit *pb.MintCredit
	if bc.MintCredit != nil {
		pbMintCredit = bc.MintCredit.ToProto().(*pb.MintCredit)
	}

	return &pb.BurnCredit{
		MintCreditID:  bc.MintCreditID,
		MintCredit:    pbMintCredit,
		BurnQuantity:  bc.BurnQuantity,
		BurnMult:      bc.BurnMult,
		BurnTimestamp: bc.BurnTimeStamp,
		Adjusted:      bc.Adjusted,
	}
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
	id := bc.GetID()
	copyBc := *bc // create a copy to avoid modifying the original object
	copyBc.MintCredit = nil
	if err := state.PutStateWithCompositeKey(stub, BURN_CREDIT_PREFIX, id, &copyBc); err != nil {
		return fmt.Errorf("could not put burn credit in state: %v", err)
	}

	// Ensure the associated MintCredit is also stored in the world state
	if bc.MintCredit != nil {
		if err := bc.MintCredit.ToWorldState(stub); err != nil {
			return fmt.Errorf("could not put associated MintCredit in state: %v", err)
		}
	}

	return nil
}

func (bc *BurnCredit) GetID() *[][]string {
	// ID is MintCreditID + BurnTimeStamp
	id := make([]string, len(bc.MintCreditID))
	copy(id, bc.MintCreditID)
	id = append(id, bc.BurnTimeStamp)
	return &[][]string{id}
}

// BurnNominalQuantity handles the initial retirement of carbon units without applying private multipliers.
func BurnNominalQuantity(stub shim.ChaincodeStubInterface, mintCreditID []string, burnQuantity int64) (*BurnCredit, error) {
	mc, err := loadMintCreditAndPerformChecks(stub, burnQuantity, mintCreditID)
	if err != nil {
		return nil, err
	}

	burnTimestamp := utils.UnixMillisNowFromStub(stub)

	bc := &BurnCredit{
		BurnQuantity:  burnQuantity,
		MintCreditID:  mintCreditID,
		MintCredit:    mc,
		BurnTimeStamp: burnTimestamp,
		Adjusted:      false,
		BurnMult:      0, // Not yet calculated
	}

	bc.MintCredit.Credit.Quantity -= burnQuantity

	if err = bc.ToWorldState(stub); err != nil {
		return nil, fmt.Errorf("could not save burn credit: %v", err)
	}

	return bc, nil
}

// ApplyBurnMultipliers calculates and applies multipliers based on company private data.
func ApplyBurnMultipliers(stub shim.ChaincodeStubInterface, burnCreditID []string) error {
	bc := &BurnCredit{
		MintCreditID: burnCreditID[:len(burnCreditID)-1],
	}
	bc.BurnTimeStamp = burnCreditID[len(burnCreditID)-1]

	if err := bc.FromWorldState(stub, burnCreditID); err != nil {
		return fmt.Errorf("could not load burn credit: %v", err)
	}

	if bc.Adjusted {
		return fmt.Errorf("multipliers already applied to this burn credit")
	}

	if identities.GetID(stub) != bc.MintCredit.OwnerID {
		return fmt.Errorf("only the owner of the credit can apply burn multipliers: %s != %s", identities.GetID(stub), bc.MintCredit.OwnerID)
	}

	// Load pseudonym to company ID mapping (private data)
	pseudonymToCompanyID := companies.PseudonymToCompanyID{
		Pseudonym: bc.MintCredit.OwnerID,
	}
	err := pseudonymToCompanyID.FromWorldState(stub, (*pseudonymToCompanyID.GetID())[0])
	if err != nil {
		return fmt.Errorf("could not get pseudonym to company ID mapping: %v", err)
	}

	company := &companies.Company{
		ID: pseudonymToCompanyID.CompanyID,
	}
	// Load company public data
	err = company.FromWorldState(stub, (*company.GetID())[0])
	if err != nil {
		return fmt.Errorf("could not get company from world state: %v", err)
	}

	pApplier := policies.NewPolicyApplier()
	pInput := &policies.PolicyInput{
		Company: company,
		Chunk:   bc.MintCredit.Chunk,
	}
	activePols, err := policies.GetActivePolicies(stub)
	if err != nil {
		return fmt.Errorf("could not get active policies: %v", err)
	}

	bc.BurnMult, err = pApplier.BurnIndependentMult(pInput, activePols)
	if err != nil {
		return fmt.Errorf("could not get burn multiplier: %v", err)
	}

	bc.Adjusted = true

	// Burn the extra quantity from the multiplier
	// BurnMult is the additional multiplier (e.g., 500 means +0.5x, total 1.5x)
	if bc.BurnMult > 0 {
		extraQuantity := (bc.BurnQuantity * bc.BurnMult) / policies.MULTIPLIER_SCALE
		if extraQuantity > 0 {
			if extraQuantity > bc.MintCredit.Quantity {
				return fmt.Errorf("not enough remaining credits to apply multiplier: need %d more, have %d", extraQuantity, bc.MintCredit.Quantity)
			}
			bc.MintCredit.Quantity -= extraQuantity
		}
	}

	// Save the adjusted BurnCredit and updated MintCredit
	if err := bc.ToWorldState(stub); err != nil {
		return fmt.Errorf("could not save adjusted burn credit and updated mint credit: %v", err)
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
