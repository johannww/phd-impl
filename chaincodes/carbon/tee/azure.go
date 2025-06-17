package tee

import (
	"fmt"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	report_verifier "github.com/johannww/phd-impl/tee_auction/go/report"
)

const (
	INITIAL_TEE_REPORT = "initialTEEReport"
)

func InitialReportToWorldState(stub shim.ChaincodeStubInterface, report *attest.SNPAttestationReport) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	reportBytes, err := report.SerializeReport()
	if err != nil {
		return fmt.Errorf("could not serialize TEE report: %v", err)
	}

	err = stub.PutState(INITIAL_TEE_REPORT, reportBytes)
	if err != nil {
		return fmt.Errorf("could not store initial TEE report: %v", err)
	}

	return nil
}
