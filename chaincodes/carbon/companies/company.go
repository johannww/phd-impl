package companies

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	"github.com/johannww/phd-impl/chaincodes/carbon/state"
)

const COMPANY_PREFIX = "company"

// Company represent a company stored as private data in the world state.
type Company struct {
	ID         string                  // ID might be the CNPJ (brazilian company national ID)
	Coordinate properties.Coordinate   // Geographical coordinate in floating point format
	DataProps  []*data.ValidationProps // How data from the company is validated
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
