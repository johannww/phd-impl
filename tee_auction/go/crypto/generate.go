package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

func CreateCertAndPrivKey(certFileName, privKeyFileName string) (
	certDERBytes []byte,
	priv ed25519.PrivateKey,
) {
	var pub ed25519.PublicKey
	var err error

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
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, pub, priv)
	panicOnError(err)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDERBytes,
	})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	panicOnError(err)
	privPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	// os.WriteFile("server.crt", pemBytes, 0644)
	// os.WriteFile("server.key", privPemBytes, 0644)
	err = os.WriteFile(certFileName, pemBytes, 0644)
	panicOnError(err)
	err = os.WriteFile(privKeyFileName, privPemBytes, 0644)
	panicOnError(err)

	return
}

func panicOnError(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}
