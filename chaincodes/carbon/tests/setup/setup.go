package setup_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	mathrand "math/rand/v2"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/attrmgr"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/johannww/phd-impl/chaincodes/carbon/identities"
	mocks "github.com/johannww/phd-impl/chaincodes/carbon/state/mocks"
)

type MockIdentities map[string][]byte

const (
	REGULAR_ID = "regular"
	IDEMIX_ID  = "idemix"
)

const (
	X509_TYPE   = "x509"
	IDEMIX_TYPE = "idemix"
)

const (
	IDEMIX_NYM_LEN = 16
)

func generateIdemix(mspName string) []byte {
	roleIdentifier, _ := proto.Marshal(&msp.MSPRole{
		Role: msp.MSPRole_CLIENT,
	})
	ouIdentifier, _ := proto.Marshal(&msp.OrganizationUnit{
		OrganizationalUnitIdentifier: mspName,
	})

	// each nym is 16 bytes long
	var nymX, nymY []byte
	for i := 0; i < IDEMIX_NYM_LEN; i++ {
		nymX = append(nymX, byte(mathrand.IntN(256)))
		nymY = append(nymY, byte(mathrand.IntN(256)))
	}

	idemixID, _ := proto.Marshal(&msp.SerializedIdemixIdentity{
		NymX:  nymX,
		NymY:  nymY,
		Role:  roleIdentifier,
		Ou:    ouIdentifier,
		Proof: []byte("proof"),
	})
	return idemixID
}

func generateX509(attrs *attrmgr.Attributes, mspName, cn string) []byte {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	keyUsage := x509.KeyUsageDigitalSignature
	marshaledAttr, err := json.Marshal(attrs)
	if err != nil {
		panic(fmt.Sprintf("could not marshal attributes: %s", err))
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{strings.ToUpper(mspName)},
			OrganizationalUnit: []string{mspName},
			CommonName:         cn,
		},
		ExtraExtensions: []pkix.Extension{{
			Id:       attrmgr.AttrOID,
			Critical: false,
			Value:    marshaledAttr,
		}},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Duration(365 * 24 * time.Hour)),

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Println(err.Error())
	}
	certBytesPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	// ioutil.WriteFile("cert.pem", certBytesPem, 0)
	return certBytesPem
}

func GenerateHFSerializedIdentity(idType string, attrs *attrmgr.Attributes, mspName, cn string) []byte {
	var credentialsBytes []byte

	switch idType {
	case X509_TYPE:
		credentialsBytes = generateX509(attrs, mspName, cn)
	case IDEMIX_TYPE:
		credentialsBytes = generateIdemix(mspName)
	default:
		panic(fmt.Sprintf("Unknown identity type: %s", idType))
	}

	hfSerializedID, _ := proto.Marshal(&msp.SerializedIdentity{
		Mspid:   mspName,
		IdBytes: credentialsBytes,
	})

	return hfSerializedID
}

func SetupIdentities(stub *mocks.MockStub) MockIdentities {
	mockIds := make(map[string][]byte)

	mockIds[REGULAR_ID] = GenerateHFSerializedIdentity(
		X509_TYPE,
		&attrmgr.Attributes{
			Attrs: map[string]string{},
		}, "AUCTIONEER", "auctioneer1",
	)

	mockIds[IDEMIX_ID] = GenerateHFSerializedIdentity(
		IDEMIX_TYPE,
		nil,
		"COMPANY",
		"",
	)

	// Generate a certificate for the price viewer
	mockIds[identities.PriceViewer] = GenerateHFSerializedIdentity(
		X509_TYPE,
		&attrmgr.Attributes{
			Attrs: map[string]string{
				identities.PriceViewer: "true",
			},
		}, "AUCTIONEER", "auctioneer1",
	)

	mockIds[identities.InteropRelayerAttr] = GenerateHFSerializedIdentity(
		X509_TYPE,
		&attrmgr.Attributes{
			Attrs: map[string]string{
				identities.InteropRelayerAttr: "true",
			},
		}, "AUCTIONEER", "auctioneer1",
	)

	// // Generate idmix identity for buyer
	// mockIds[IDEMIX_ID] = generateIdemix(attrmgr.Attributes{
	// 	Attrs: map[string]string{}}, "BUYER", "buyer1",
	// )

	return mockIds
}
