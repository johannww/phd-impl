package carbon_tests

import (
	"testing"

	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
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

	err := companies.CompanyToWorldState(stub, company)
	require.NoError(t, err, "Failed to register company")
}
