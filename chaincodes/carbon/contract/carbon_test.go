package contract

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/stretchr/testify/require"
)

func TestContractApiRuntimeFunctionParsing(t *testing.T) {
	_, err := contractapi.NewChaincode(NewCarbonContract())
	require.NoError(t, err, "Error parsing contract API runtime functions")
}
