package carbon_tests

import (
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
	"github.com/johannww/phd-impl/chaincodes/carbon/tee"
	setup "github.com/johannww/phd-impl/chaincodes/carbon/tests/setup"
	tee_auction "github.com/johannww/phd-impl/tee_auction/go/auction"
	tee_auction_crypto "github.com/johannww/phd-impl/tee_auction/go/crypto"
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
	mockIds := setup.SetupIdentities(stub)
	stub.Creator = mockIds[identities.TEEConfigurer]

	err = tee.ExpectedCCEPolicyToWorldState(stub, ccePolicyB64)
	require.NoError(t, err, "Failed to store CCE policy in world state")

	err = tee.InitialReportToWorldState(stub, reportJsonBytes)
	require.NoError(t, err, "Failed to store initial TEE report in world state")

}

func TestVerifyAuctionAppSignature(t *testing.T) {
	stub := mocks.NewMockStub("carbon", nil)
	mockIds := setup.SetupIdentities(stub)
	stub.Creator = mockIds[identities.TEEConfigurer]

	certDerBytes, privKey := tee_auction_crypto.CreateCertAndPrivKey("/dev/null", "/dev/null")
	genMockReportWithCertDerHash(t, stub, certDerBytes)

	serializedResults := &tee_auction.SerializedAuctionResultTEE{
		ResultBytes:  []byte("test result bytes"),
		ReceivedHash: []byte("received hash"),
		AppSignature: []byte("wrong signature"),
		TEECertDer:   certDerBytes,
	}

	// Test with wrong hash
	stub.MockTransactionStart("tx2")
	valid, err := tee.VerifyAuctionAppSignature(stub, serializedResults.ResultBytes, serializedResults.ReceivedHash, serializedResults.AppSignature, serializedResults.TEECertDer)
	require.Error(t, err, "could not verify TEE Application signature")
	require.False(t, valid, "Expected invalid signature verification due to wrong hash and signature")
	stub.MockTransactionEnd("tx2")

	bytesSignedByTEE := []byte{}
	bytesSignedByTEE = append(bytesSignedByTEE, serializedResults.ResultBytes...)
	bytesSignedByTEE = append(bytesSignedByTEE, serializedResults.ReceivedHash...)
	signature := ed25519.Sign(privKey, bytesSignedByTEE)
	serializedResults.AppSignature = signature

	// Test with correct signature
	stub.MockTransactionStart("tx3")
	valid, err = tee.VerifyAuctionAppSignature(stub, serializedResults.ResultBytes, serializedResults.ReceivedHash, serializedResults.AppSignature, serializedResults.TEECertDer)
	require.NoError(t, err, "could not verify TEE Application signature")
	require.True(t, valid, "Expected valid signature verification")
	stub.MockTransactionEnd("tx3")

}

func genMockReportWithCertDerHash(t *testing.T, stub *mocks.MockStub, certDerBytes []byte) {
	certHash := sha512.Sum512(certDerBytes)
	certHashHex := hex.EncodeToString(certHash[:])
	report := &attest.SNPAttestationReport{
		ReportData: certHashHex,
	}
	reportJsonBytes, err := json.Marshal(report)
	require.NoError(t, err, "Failed to generate mock report with cert der hash")
	stub.MockTransactionStart("tx1")
	err = stub.PutState(tee.INITIAL_TEE_REPORT, reportJsonBytes)
	require.NoError(t, err, "Failed to store mock report with cert der hash in world state")
	stub.MockTransactionEnd("tx1")
}
