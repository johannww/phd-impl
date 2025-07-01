package carbon_tests

import (
	"encoding/json"
	"os"
	"testing"

	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/carbon/tee"
	"github.com/stretchr/testify/require"
)

func TestAzureCEEPolicyVerification(t *testing.T) {
	armTemplateJsonBytes, err := os.ReadFile("./data/azure/arm_template.json")
	require.NoError(t, err, "Failed to read ARM template JSON file")

	reportJsonBytes, err := os.ReadFile("./data/azure/report.json")
	require.NoError(t, err, "Failed to read ARM template JSON file")

	var armJson map[string]interface{}
	err = json.Unmarshal(armTemplateJsonBytes, &armJson)
	require.NoError(t, err, "Failed to unmarshal ARM template JSON")

	resources := armJson["resources"].([]interface{})
	firstResource := resources[0].(map[string]interface{})
	confidentialComputeProperties := firstResource["properties"].(map[string]interface{})["confidentialComputeProperties"].(map[string]interface{})
	ccePolicyB64 := confidentialComputeProperties["ccePolicy"].(string)

	stub := mocks.NewMockStub("carbon", nil)
	stub.MockTransactionStart("tx1")

	err = tee.ExpectedCCEPolicyToWorldState(stub, ccePolicyB64)
	require.NoError(t, err, "Failed to store CCE policy in world state")

	err = tee.InitialReportToWorldState(stub, reportJsonBytes)
	require.NoError(t, err, "Faied to store initial TEE report in world state")

}
