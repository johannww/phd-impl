package carbon_tests

import (
	"encoding/json"
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/stretchr/testify/require"
)

func TestCompanyRegister(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)

	stub.Creator = possibleIds[setup.IDEMIX_ID]
	company := &companies.Company{
		ID: "12345678901234", // Example CNPJ
		Coordinate: &utils.Coordinate{
			Latitude:  -23.550520,
			Longitude: -46.633308,
		},
		DataProps: &data.ValidationProps{
			Methods: []data.ValidationMethod{data.ValidationMethodGroundTruth},
		},
	}

	err := companies.RegisterCompany(stub, company)
	require.NoError(t, err, "Failed to register company")
}

func TestCreatePseudonymToCompanyID(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)

	stub.Creator = possibleIds[setup.IDEMIX_ID]
	stub.MockTransactionStart("tx1")

	mapping := &companies.PseudonymToCompanyID{
		Pseudonym: identities.GetID(stub),
		CompanyID: "12345678901234",
	}

	transientData, err := json.Marshal(mapping)
	require.NoError(t, err, "Failed to marshal pseudonym mapping")

	stub.TransientMap = map[string][]byte{
		companies.PSEUDONYM_TO_COMPANY_ID_TRANSIENT_KEY: transientData,
	}

	// Create the mapping
	err = companies.CreatePseudonymToCompanyID(stub)
	require.NoError(t, err, "Failed to create pseudonym mapping")

	// Retrieve and verify the mapping
	retrievedCompanyID, err := companies.GetCompanyIDByPseudonym(stub, mapping.Pseudonym)
	require.NoError(t, err, "Failed to retrieve company ID by pseudonym")
	require.Equal(t, mapping.CompanyID, retrievedCompanyID, "Retrieved company ID does not match")

	stub.MockTransactionEnd("tx1")
}
