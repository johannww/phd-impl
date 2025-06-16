package auction

import (
	"crypto/sha512"
	"fmt"

	cc_auction "github.com/johannww/phd-impl/chaincodes/carbon/auction"
	"github.com/johannww/phd-impl/tee_auction/report"
)

type AuctionResultTEE struct {
	matched
}

type SerializedAuctionResultTEE struct {
	// ResultBytes is a serialized
	ResultBytes []byte `json:"resultBytes"`
	// AmdReportBytes is a serialized &attest.SNPAttestationReport{}
	AmdReportBytes []byte `json:"reportBytes"`
}

// TODOHP: finish auction running on tee
func RunTEEAuction(auctionData *cc_auction.AuctionData) (*SerializedAuctionResultTEE, error) {
	result := &SerializedAuctionResultTEE{}

	// Validate data commtiment

	// get report on the results
	err := result.setHardwareSignature()
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (result *SerializedAuctionResultTEE) setHardwareSignature() error {
	var err error
	reportUserData := sha512.Sum512(result.ResultBytes)

	result.AmdReportBytes, err = report.GetAmdSevSnpReport(reportUserData)
	if err != nil {
		return fmt.Errorf("Failed to get AMD SEV SNP report: %v", err)
	}

	return nil
}
