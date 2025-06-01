package report

import (
	"fmt"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
)

const USER_DATA_SIZE = attest.REPORT_DATA_SIZE

func GetAmdSevSnpReport(reportUserData [USER_DATA_SIZE]byte) ([]byte, error) {
	reportFetcher, err := attest.NewAttestationReportFetcher()
	if err != nil {
		return nil, fmt.Errorf("Failed to create attestation report fetcher: %v", err)
	}

	report, err := reportFetcher.FetchAttestationReportByte(reportUserData)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch attestation report: %v", err)
	}

	return report, nil

}
