package companies

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const PSEDONYM_TO_COMPANY_ID_PREFIX = "psedonymToCompanyID"

type PsedonymToCompanyID struct {
	Psedonym  string `json:"psedonym"`  // Pseudonym of the company (provided by identities.GetID(stub))
	CompanyID string `json:"companyID"` // ID of the company, e.g., CNPJ in Brazil
}

var _ state.WorldStateManager = (*PsedonymToCompanyID)(nil)

// FromWorldState implements state.WorldStateManager.
func (p *PsedonymToCompanyID) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetPvtDataWithCompositeKey(stub, PSEDONYM_TO_COMPANY_ID_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, p)
}

// GetID implements state.WorldStateManager.
func (p *PsedonymToCompanyID) GetID() *[][]string {
	return &[][]string{{p.Psedonym}}
}

// ToWorldState implements state.WorldStateManager.
func (p *PsedonymToCompanyID) ToWorldState(stub shim.ChaincodeStubInterface) error {
	firstID := (*p.GetID())[0]
	return state.PutPvtDataWithCompositeKey(stub, PSEDONYM_TO_COMPANY_ID_PREFIX, firstID, state.BIDS_PVT_DATA_COLLECTION, p)
}
