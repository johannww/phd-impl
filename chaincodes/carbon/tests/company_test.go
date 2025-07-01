package carbon_tests

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	"github.com/stretchr/testify/require"
)

func TestCompanyRegister(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	possibleIds := setup.SetupIdentities(stub)

	stub.Creator = possibleIds[setup.IDEMIX_ID]
	company := &companies.Company{
		ID: "12345678901234", // Example CNPJ
		Coordinate: &properties.Coordinate{
			Latitude:  -23.550520,
			Longitude: -46.633308,
		},
		DataProps: &data.ValidationProps{
			Methods: []data.ValidationMethod{data.ValidationMethodGroundTruth},
		},
	}

	err := companies.CompanyToWorldState(stub, company)
	require.NoError(t, err, "Failed to register company")
}
