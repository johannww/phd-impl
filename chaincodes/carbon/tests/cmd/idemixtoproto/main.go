package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"flag"
	"log"

	"github.com/IBM/idemix/idemixmsp"
	"github.com/golang/protobuf/proto"
)

// from https://github.com/hyperledger/fabric-ca/blob/b3ae5fc317baf5306aa690bbab4113c78e606e3d/lib/client/credential/idemix/signerconfig.go#L10-L27
type SignerConfig struct {
	// Cred represents the serialized idemix credential of the default signer
	Cred []byte `protobuf:"bytes,1,opt,name=Cred,proto3" json:"Cred,omitempty"`
	// Sk is the secret key of the default signer, corresponding to credential Cred
	Sk []byte `protobuf:"bytes,2,opt,name=Sk,proto3" json:"Sk,omitempty"`
	// OrganizationalUnitIdentifier defines the organizational unit the default signer is in
	OrganizationalUnitIdentifier string `protobuf:"bytes,3,opt,name=organizational_unit_identifier,json=organizationalUnitIdentifier" json:"organizational_unit_identifier,omitempty"`
	// Role defines whether the default signer is admin, member, peer, or client
	Role int `protobuf:"varint,4,opt,name=role,json=role" json:"role,omitempty"`
	// EnrollmentID contains the enrollment id of this signer
	EnrollmentID string `protobuf:"bytes,5,opt,name=enrollment_id,json=enrollmentId" json:"enrollment_id,omitempty"`
	// CRI contains a serialized Credential Revocation Information
	CredentialRevocationInformation []byte `protobuf:"bytes,6,opt,name=credential_revocation_information,json=credentialRevocationInformation,proto3" json:"credential_revocation_information,omitempty"`
	// CurveID specifies the name of the Idemix curve to use, defaults to 'amcl.Fp256bn'
	CurveID string `protobuf:"bytes,7,opt,name=curve_id,json=curveID" json:"curveID,omitempty"`
	// RevocationHandle is the handle used to single out this credential and determine its revocation status
	RevocationHandle string `protobuf:"bytes,8,opt,name=revocation_handle,json=revocationHandle,proto3" json:"revocation_handle,omitempty"`
}

func toIdemixMSPSignerConfig(sc *SignerConfig) *idemixmsp.IdemixMSPSignerConfig {
	return &idemixmsp.IdemixMSPSignerConfig{
		Cred:                            sc.Cred,
		Sk:                              sc.Sk,
		OrganizationalUnitIdentifier:    sc.OrganizationalUnitIdentifier,
		Role:                            int32(sc.Role),
		EnrollmentId:                    sc.EnrollmentID,
		CredentialRevocationInformation: sc.CredentialRevocationInformation,
		RevocationHandle:                sc.RevocationHandle,
		CurveId:                         sc.CurveID,
		Schema:                          "",
	}
}

func ConvertJSONToProto(jsonData []byte) ([]byte, error) {
	var sc SignerConfig
	err := json.Unmarshal(jsonData, &sc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	idConfig := toIdemixMSPSignerConfig(&sc)

	protoBytes, err := proto.Marshal(idConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proto: %w", err)
	}

	return protoBytes, nil
}

func main() {
	inputFiles := flag.String("input", "", "Comma-separated list of input JSON files")
	flag.Parse()

	if *inputFiles == "" {
		log.Fatal("Input file paths must be provided")
	}

	inputList := splitAndTrim(*inputFiles)

	for _, inputFile := range inputList {
		jsonData, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatalf("Failed to read input file %s: %v", inputFile, err)
		}

		protoBytes, err := ConvertJSONToProto(jsonData)
		if err != nil {
			log.Fatalf("Conversion failed for file %s: %v", inputFile, err)
		}

		outputDir := fmt.Sprintf("%s/user", getUserRoot(inputFile))
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create output directory %s: %v", outputDir, err)
		}

		outputFile := fmt.Sprintf("%s/%s", outputDir, getBaseName(inputFile))

		err = os.WriteFile(outputFile, protoBytes, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file %s: %v", outputFile, err)
		}
	}
}

func splitAndTrim(s string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func getUserRoot(filePath string) string {
	dir := path.Dir(filePath)
	return path.Join(dir, "..", "..")
}

func getBaseName(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}
