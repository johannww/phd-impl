package main

import (
	"crypto/sha512"
	"fmt"
	"net/http"

	"github.com/johannww/phd-impl/tee_auction/go/api"
	"github.com/johannww/phd-impl/tee_auction/go/crypto"
	"github.com/johannww/phd-impl/tee_auction/go/report"
)

// TODO: This program will execute the auction on an
// Azure Confidential Container.
func main() {

	// Generate self-signed certificate and private key
	certFileName := "server.crt"
	privKeyFileName := "server.key"
	certDERBytes, priv := crypto.CreateCertAndPrivKey(certFileName, privKeyFileName)

	// Put the public key's hash in the AMD SEV-SNP report
	// on ReportData field. See:
	// https://github.com/johannww/ubuntu-learning/blob/4933313b537c6283dbef8eb093b3924611c5061c/crypto/azure_tee/containers/report_attester/snp_attestation_report.go#L39-L40
	// See page 56 (REPORT_DATA) of https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
	reportUserData := [report.USER_DATA_SIZE]byte{}
	certHash := sha512.Sum512(certDERBytes) // Hash the certificate bytes
	for i := 0; i < len(certHash); i++ {
		reportUserData[i] = certHash[i]
	}

	reportBytes, err := report.GetAmdSevSnpReport(reportUserData)
	if err != nil {
		panic(fmt.Sprintf("Failed to get AMD SEV-SNP report: %v", err))
	}
	fmt.Printf("AMD SEV-SNP report: %x\n", reportBytes)

	// Wait for requests
	auctionServer := &api.AuctionServer{
		ReportBytes: reportBytes,
	}
	router := auctionServer.SetupRouter(priv, certDERBytes)
	// err = http.ListenAndServe(":8080", router)
	err = http.ListenAndServeTLS(":8080", certFileName, privKeyFileName, router)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
