package main

import (
	"crypto/ed25519"
	"crypto/rand"
)

// TODO: This program will execute the auction on an
// Azure Confidential Container.
func main() {

	// Generate key pair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	if len(pub) != ed25519.PublicKeySize {
		panic("Invalid public key size")
	}

	_ = priv

	// Put the public key's hash in the AMD SEV-SNP report
	// on ReportData field. See:
	// https://github.com/johannww/ubuntu-learning/blob/4933313b537c6283dbef8eb093b3924611c5061c/crypto/azure_tee/containers/report_attester/snp_attestation_report.go#L39-L40
	// See page 56 (REPORT_DATA) of https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf

	// Wait for requests

}
