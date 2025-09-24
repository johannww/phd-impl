package companies

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
	"github.com/johannww/phd-impl/chaincodes/carbon/utils"
)

const COMPANY_PREFIX = "company"

// Company represent a company stored as private data in the world state.
type Company struct {
	ID         string                // ID might be the CNPJ (brazilian company national ID)
	Coordinate *utils.Coordinate     // Geographical coordinate in floating point format
	DataProps  *data.ValidationProps // How data from the company is validated
}

var _ state.WorldStateManager = (*Company)(nil)

// FromWorldState implements state.WorldStateManager.
func (c *Company) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetPvtDataWithCompositeKey(stub, COMPANY_PREFIX, keyAttributes, state.BIDS_PVT_DATA_COLLECTION, c)
}

// GetID implements state.WorldStateManager.
func (c *Company) GetID() *[][]string {
	return &[][]string{{c.ID}}
}

// ToWorldState implements state.WorldStateManager.
func (c *Company) ToWorldState(stub shim.ChaincodeStubInterface) error {
	firstID := (*c.GetID())[0]
	return state.PutPvtDataWithCompositeKey(stub, COMPANY_PREFIX, firstID, state.BIDS_PVT_DATA_COLLECTION, c)
}

func CompanyToWorldState(stub shim.ChaincodeStubInterface, company *Company) error {
	if company == nil {
		return nil
	}
	if company.ID == "" {
		return fmt.Errorf("company ID cannot be empty")
	}
	if *company.Coordinate == (utils.Coordinate{}) {
		return fmt.Errorf("company coordinate cannot be empty")
	}
	if company.DataProps == nil || len(company.DataProps.Methods) == 0 {
		return fmt.Errorf("company data properties cannot be empty")
	}

	return company.ToWorldState(stub)
}
