package tee

import (
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
	report := attest.SNPAttestationReport{}
	err := json.Unmarshal(reportJsonBytes, &report)
	if err != nil {
		return fmt.Errorf("could not unmarshal TEE report: %v", err)
	}

	verifies, err := report_verifier.VerifyReportSignature(reportJsonBytes)
	if err != nil || !verifies {
		return fmt.Errorf("could not verify TEE report signature: %v", err)
	}

	err = stub.PutState(INITIAL_TEE_REPORT, reportJsonBytes)
	if err != nil {
		return fmt.Errorf("could not store initial TEE report: %v", err)
	}

	return nil
}
