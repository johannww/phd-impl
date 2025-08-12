package companies

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const PSEUDONYM_TO_COMPANY_ID_PREFIX = "pseudonymToCompanyID"

type PseudonymToCompanyID struct {
	Pseudonym string `json:"pseudonym"` // Pseudonym of the company (provided by identities.GetID(stub))
	CompanyID string `json:"companyID"` // ID of the company, e.g., CNPJ in Brazil
}

var _ state.WorldStateManager = (*PseudonymToCompanyID)(nil)

// FromWorldState implements state.WorldStateManager.
func (p *PseudonymToCompanyID) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetPvtDataWithCompositeKey(stub, PSEUDONYM_TO_COMPANY_ID_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, p)
}

// GetID implements state.WorldStateManager.
func (p *PseudonymToCompanyID) GetID() *[][]string {
	return &[][]string{{p.Pseudonym}}
}

// ToWorldState implements state.WorldStateManager.
func (p *PseudonymToCompanyID) ToWorldState(stub shim.ChaincodeStubInterface) error {
	firstID := (*p.GetID())[0]
	return state.PutPvtDataWithCompositeKey(stub, PSEUDONYM_TO_COMPANY_ID_PREFIX, firstID, state.BIDS_PVT_DATA_COLLECTION, p)
}

// CreatePseudonymToCompanyID creates a new mapping and saves it to the world state.
func CreatePseudonymToCompanyID(stub shim.ChaincodeStubInterface, pseudonym, companyID string) error {
	pseudonymToCompanyID := &PseudonymToCompanyID{
		Pseudonym: pseudonym,
		CompanyID: companyID,
	}
	return pseudonymToCompanyID.ToWorldState(stub)
}

// GetCompanyIDByPseudonym returns the company ID for a given pseudonym.
func GetCompanyIDByPseudonym(stub shim.ChaincodeStubInterface, pseudonym string) (string, error) {
	pseudonymToCompanyID := &PseudonymToCompanyID{}
	err := pseudonymToCompanyID.FromWorldState(stub, []string{pseudonym})
	if err != nil {
		return "", err
	}
	return pseudonymToCompanyID.CompanyID, nil
}
