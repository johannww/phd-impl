package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/johannww/phd-impl/tee_auction/go/api"
	"github.com/johannww/phd-impl/tee_auction/go/report"
)

// TODO: This program will execute the auction on an
// Azure Confidential Container.
func main() {

	// Generate self-signed certificate and private key
	certFileName, privKeyFileName, certBytes, priv := CreateCertAndPrivKey()

	// Put the public key's hash in the AMD SEV-SNP report
	// on ReportData field. See:
	// https://github.com/johannww/ubuntu-learning/blob/4933313b537c6283dbef8eb093b3924611c5061c/crypto/azure_tee/containers/report_attester/snp_attestation_report.go#L39-L40
	// See page 56 (REPORT_DATA) of https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
	reportUserData := [report.USER_DATA_SIZE]byte{}
	certHash := sha512.Sum512(certBytes) // Hash the certificate bytes
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
	router := auctionServer.SetupRouter(priv)
	// err = http.ListenAndServe(":8080", router)
	err = http.ListenAndServeTLS(":8080", certFileName, privKeyFileName, router)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}

func CreateCertAndPrivKey() (certFileName, privKeyFileName string,
	certBytes []byte,
	priv ed25519.PrivateKey,
) {
	var pub ed25519.PublicKey
	var err error
	certFileName = "server.crt"
	privKeyFileName = "server.key"

	// Generate key pair
	pub, priv, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	// Generate self-signed certificate
	// TODO: review this generation
	certTemplate := x509.Certificate{
		SerialNumber:          nil, // Use default serial number
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for one year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		IsCA:                  true, // Self-signed certificate
		PublicKeyAlgorithm:    x509.Ed25519,
		PublicKey:             pub,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, pub, priv)
	panicOnError(err)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	privPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: priv,
	})

	// os.WriteFile("server.crt", pemBytes, 0644)
	// os.WriteFile("server.key", privPemBytes, 0644)
	err = os.WriteFile(certFileName, pemBytes, 0644)
	panicOnError(err)
	err = os.WriteFile(privKeyFileName, privPemBytes, 0644)
	panicOnError(err)

	return
}
