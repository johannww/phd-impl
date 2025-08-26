package tee

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	tee_auction "github.com/johannww/phd-impl/tee_auction/go/auction"
	report_verifier "github.com/johannww/phd-impl/tee_auction/go/report"
)

const (
	INITIAL_TEE_REPORT = "initialTEEReport"
	CCE_POLICY         = "ccePolicy"
)

func InitialReportToWorldState(stub shim.ChaincodeStubInterface, reportJsonBytes []byte) error {
	report := attest.SNPAttestationReport{}
	err := json.Unmarshal(reportJsonBytes, &report)
	if err != nil {
		return fmt.Errorf("could not unmarshal initial TEE report: %v", err)
	}

	verifies, err := report_verifier.VerifyReportSignature(&report)
	if err != nil || !verifies {
		return fmt.Errorf("could not verify TEE report signature: %v", err)
	}

	err = verifyCCEPolicy(stub, report)
	if err != nil {
		return fmt.Errorf("could not verify CCE policy: %v", err)
	}

	err = stub.PutState(INITIAL_TEE_REPORT, reportJsonBytes)
	if err != nil {
		return fmt.Errorf("could not store initial TEE report: %v", err)
	}

	return nil
}

func VerifyAuctionResultReportSignature(
	auctionReportJsonBytes []byte,
	expectedResult []byte,
	hashReceivedByTEE []byte,
) (bool, error) {
	bytesHashedByTEE := []byte{}
	bytesHashedByTEE = append(bytesHashedByTEE, expectedResult...)
	bytesHashedByTEE = append(bytesHashedByTEE, hashReceivedByTEE...)
	expectedResultHash := sha512.Sum512(bytesHashedByTEE)

	report := attest.SNPAttestationReport{}
	err := json.Unmarshal(auctionReportJsonBytes, &report)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal auction report: %v", err)
	}

	reportDataBytes, err := hex.DecodeString(report.ReportData)
	if err != nil {
		return false, fmt.Errorf("could not decode report data string as hex: %v", err)
	}

	if bytes.Equal(reportDataBytes, expectedResultHash[:]) {
		return false, fmt.Errorf("report data does not match expected result hash")
	}

	verifies, err := report_verifier.VerifyReportSignatureJsonBytes(auctionReportJsonBytes)
	if err != nil || !verifies {
		return false, fmt.Errorf("could not verify TEE report signature: %v", err)
	}

	return true, nil
}

// TODO: Test
func VerifyAuctionAppSignature(
	stub shim.ChaincodeStubInterface,
	resultBytes []byte,
	hashReceivedByTEE []byte,
	signatureBytes []byte,
	certDer []byte,
) (bool, error) {
	initialReportBytes, err := stub.GetState(INITIAL_TEE_REPORT)
	if err != nil {
		return false, fmt.Errorf("could not get initial TEE report: %v", err)
	}
	report := attest.SNPAttestationReport{}
	err = json.Unmarshal(initialReportBytes, &report)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal initial report: %v", err)
	}

	reportDataBytes, err := hex.DecodeString(report.ReportData)
	if err != nil {
		return false, fmt.Errorf("could not decode report data string as hex: %v", err)
	}

	certHash := sha512.Sum512(certDer)
	if bytes.Equal(reportDataBytes, certHash[:]) {
		return false, fmt.Errorf("report data does not match SHA512 of the provided certificate")
	}

	cert, err := x509.ParseCertificate(certDer)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %v", err)
	}

	bytesSignedByTEE := []byte{}
	bytesSignedByTEE = append(bytesSignedByTEE, resultBytes...)
	bytesSignedByTEE = append(bytesSignedByTEE, hashReceivedByTEE...)

	verifies := ed25519.Verify(
		cert.PublicKey.(ed25519.PublicKey),
		bytesSignedByTEE,
		signatureBytes)
	if !verifies {
		return false, fmt.Errorf("could not verify TEE Application signature: %v", err)
	}

	return true, nil
}

func verifyCCEPolicy(stub shim.ChaincodeStubInterface, report attest.SNPAttestationReport) error {
	ccePolicyBase64, err := GetCCEPolicy(stub)
	if err != nil {
		return fmt.Errorf("could not get CCE policy: %v", err)
	}

	ccePoplicyBytes, err := base64.StdEncoding.DecodeString(ccePolicyBase64)
	if err != nil {
		return fmt.Errorf("could not decode CCE policy from hex: %v", err)
	}

	ccePolciySha256FromWorldState := sha256.Sum256(ccePoplicyBytes)

	ccePolicySha256HashFromReport, err := hex.DecodeString(report.HostData) // see https://www.youtube.com/watch?v=H9DP5CMqGac
	if err != nil {
		return fmt.Errorf("could not unmarshal CCE policy: %v", err)
	}

	if !bytes.Equal(ccePolicySha256HashFromReport, ccePolciySha256FromWorldState[:]) {
		return fmt.Errorf("CCE policy hash from report does not match the expected CCE policy hash")
	}

	return nil
}

// ExpectedCCEPolicyToWorldState stores the expected CCE policy in the world state.
// In the azure ARM template, the CCE policy is base64 and stays in the json path:
// .resources[0].properties.confidentialComputeProperties.ccePolicy
// The CCE policy contains information about the docker image being run in the TEE.
// From the CCE policy, it is possible to identify the docker image.
func ExpectedCCEPolicyToWorldState(stub shim.ChaincodeStubInterface, base64CcePolicy string) error {
	if err := stub.PutState(CCE_POLICY, []byte(base64CcePolicy)); err != nil {

		return fmt.Errorf("could not store CCE policy: %v", err)
	}
	return nil
}

func GetCCEPolicy(stub shim.ChaincodeStubInterface) (string, error) {
	ccePolicy, err := stub.GetState(CCE_POLICY)
	if err != nil {
		return "", fmt.Errorf("could not get CCE policy: %v", err)
	}
	if ccePolicy == nil {
		return "", fmt.Errorf("CCE policy not found in world state")
	}
	return string(ccePolicy), nil
}

func VerifyTEEResult(stub shim.ChaincodeStubInterface, serializedResults *tee_auction.SerializedAuctionResultTEE) error {
	verifies, err := VerifyAuctionResultReportSignature(serializedResults.AmdReportBytes,
		serializedResults.ResultBytes,
		serializedResults.ReceivedHash)
	if err != nil {
		return fmt.Errorf("could not verify TEE auction result report signature: %v", err)
	}
	if !verifies {
		return fmt.Errorf("TEE auction result report signature is invalid")
	}

	verifies, err = VerifyAuctionAppSignature(stub,
		serializedResults.ResultBytes,
		serializedResults.ReceivedHash,
		serializedResults.AppSignature,
		serializedResults.TEECertDer)
	if err != nil {
		return fmt.Errorf("could not verify TEE auction app signature: %v", err)
	}
	if !verifies {
		return fmt.Errorf("TEE auction app signature is invalid")
	}

	return nil
}
