package tee

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
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
	expectedResultHash []byte) (bool, error) {
	report := attest.SNPAttestationReport{}
	err := json.Unmarshal(auctionReportJsonBytes, &report)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal auction report: %v", err)
	}

	reportDataBytes, err := hex.DecodeString(report.ReportData)
	if err != nil {
		return false, fmt.Errorf("could not decode report data string as hex: %v", err)
	}

	if bytes.Compare(reportDataBytes, expectedResultHash) != 0 {
		return false, fmt.Errorf("report data does not match expected result hash")
	}

	verifies, err := report_verifier.VerifyReportSignatureJsonBytes(auctionReportJsonBytes)
	if err != nil || !verifies {
		return false, fmt.Errorf("could not verify TEE report signature: %v", err)
	}

	return true, nil
}

// TODO: test this
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

	if bytes.Compare(ccePolicySha256HashFromReport, ccePolciySha256FromWorldState[:]) != 0 {
		return fmt.Errorf("CCE policy hash from report does not match the expected CCE policy hash")
	}

	return nil
}

func CCEPolicyToWorldState(stub shim.ChaincodeStubInterface, base64CcePolicy string) error {
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
