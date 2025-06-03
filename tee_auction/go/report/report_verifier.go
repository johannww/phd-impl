package report

import (
	"crypto/ecdsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"slices"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
)

func VerifyReportSignature(reportJsonBytes []byte) (bool, error) {
	report := attest.SNPAttestationReport{}
	err := json.Unmarshal(reportJsonBytes, &report)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal attestation report: %v", err)
	}

	certFetcher := attest.DefaultAMDMilanCertFetcherNew()
	chain, anInt, error := certFetcher.GetCertChain(report.ChipID, report.ReportedTCB)
	if error != nil {
		fmt.Printf("Error: %s\n", error)
		fmt.Printf("Error: %d\n", anInt)
	}
	// fmt.Printf("chain: %+v\n", chain)
	// os.WriteFile("cert_chain.pem", chain, 0o644)

	rest := []byte{}
	derCerts := []byte{}

	certPemBlock, rest := pem.Decode(chain)
	derCerts = append(derCerts, certPemBlock.Bytes...)
	for len(rest) > 0 {
		certPemBlock, restt := pem.Decode(rest)
		rest = restt
		derCerts = append(derCerts, certPemBlock.Bytes...)
	}

	x509Certs, err := x509.ParseCertificates(derCerts)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	attestationReportBytes, err := report.SerializeReport()
	if err != nil {
		return false, fmt.Errorf("Failed to serialize attestation report: %v", err)
	}

	hash := sha512.New384()
	hash.Write(attestationReportBytes[:672])
	fmt.Println(len(attestationReportBytes))
	sum := hash.Sum(nil)
	pubKey := x509Certs[0].PublicKey.(*ecdsa.PublicKey)

	// see sig spec: https: //www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
	reportSigBytes := attestationReportBytes[672:]
	rBigEndian := reportSigBytes[:72]
	fmt.Printf("rLittleEndian: %x\n", rBigEndian)
	slices.Reverse[[]byte](rBigEndian)
	fmt.Printf("rBigEndian: %x\n", rBigEndian)
	r := new(big.Int).SetBytes(rBigEndian) // set bytes treat as big endian
	sBigEndian := reportSigBytes[72:144]
	fmt.Printf("sLittleEndian: %x\n", sBigEndian)
	slices.Reverse[[]byte](sBigEndian)
	s := new(big.Int).SetBytes(sBigEndian) // set bytes treat as big endian
	fmt.Printf("sBigEndian: %x\n", sBigEndian)

	sigValid := ecdsa.Verify(pubKey, sum, r, s)
	fmt.Printf("sigValid: %t\n", sigValid)

	return sigValid, nil

}

func VerifyReportSignatureBase64(reportBase64Bytes []byte) (bool, error) {
	attestationReportBytes, err := base64.StdEncoding.DecodeString(string(reportBase64Bytes))
	if err != nil {
		return false, fmt.Errorf("Failed to decode base64 attestation report: %v", err)
	}

	report := attest.SNPAttestationReport{}
	err = report.DeserializeReport(attestationReportBytes)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal attestation report: %v", err)
	}

	certFetcher := attest.DefaultAMDMilanCertFetcherNew()
	chain, anInt, error := certFetcher.GetCertChain(report.ChipID, report.ReportedTCB)
	if error != nil {
		fmt.Printf("Error: %s\n", error)
		fmt.Printf("Error: %d\n", anInt)
	}
	// fmt.Printf("chain: %+v\n", chain)
	// os.WriteFile("cert_chain.pem", chain, 0o644)

	rest := []byte{}
	derCerts := []byte{}

	certPemBlock, rest := pem.Decode(chain)
	derCerts = append(derCerts, certPemBlock.Bytes...)
	for len(rest) > 0 {
		certPemBlock, restt := pem.Decode(rest)
		rest = restt
		derCerts = append(derCerts, certPemBlock.Bytes...)
	}

	x509Certs, err := x509.ParseCertificates(derCerts)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	hash := sha512.New384()
	hash.Write(attestationReportBytes[:672])
	fmt.Println(len(attestationReportBytes))
	sum := hash.Sum(nil)
	pubKey := x509Certs[0].PublicKey.(*ecdsa.PublicKey)

	// see sig spec: https: //www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
	reportSigBytes := attestationReportBytes[672:]
	rBigEndian := reportSigBytes[:72]
	fmt.Printf("rLittleEndian: %x\n", rBigEndian)
	slices.Reverse[[]byte](rBigEndian)
	fmt.Printf("rBigEndian: %x\n", rBigEndian)
	r := new(big.Int).SetBytes(rBigEndian) // set bytes treat as big endian
	sBigEndian := reportSigBytes[72:144]
	fmt.Printf("sLittleEndian: %x\n", sBigEndian)
	slices.Reverse[[]byte](sBigEndian)
	s := new(big.Int).SetBytes(sBigEndian) // set bytes treat as big endian
	fmt.Printf("sBigEndian: %x\n", sBigEndian)

	sigValid := ecdsa.Verify(pubKey, sum, r, s)
	fmt.Printf("sigValid: %t\n", sigValid)

	return sigValid, nil

}
