package tee

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	report_verifier "github.com/johannww/phd-impl/tee_auction/go/report"
)

const (
	INITIAL_TEE_REPORT = "initialTEEReport"
)

func InitialReportToWorldState(stub shim.ChaincodeStubInterface, reportJsonBytes []byte) error {
	verifies, err := report_verifier.VerifyReportSignatureJsonBytes(reportJsonBytes)
	if err != nil || !verifies {
		return fmt.Errorf("could not verify TEE report signature: %v", err)
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
