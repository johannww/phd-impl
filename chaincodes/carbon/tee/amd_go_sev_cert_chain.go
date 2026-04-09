package tee

import (
	"fmt"

	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const SEV_CERT_CHAIN_PREFIX = "sevCertChain"

// SevCertChain wraps the sevsnp.CertificateChain and implements WorldStateManager
// to enable persistent storage of AMD SEV-SNP certificate chains in the world state.
// This allows caching of the certificate chain to avoid repeated fetches from AMD KDS.
type SevCertChain struct {
	*sevsnp.CertificateChain
}

// FromProto implements [state.WorldStateManager].
func (s *SevCertChain) FromProto(m proto.Message) error {
	cc, ok := m.(*sevsnp.CertificateChain)
	if !ok {
		return fmt.Errorf("unexpected proto message type for SevCertChain")
	}
	s.CertificateChain = cc
	return nil
}

// FromWorldState implements [state.WorldStateManager].
func (s *SevCertChain) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, SEV_CERT_CHAIN_PREFIX, keyAttributes, s)
}

// GetID implements [state.WorldStateManager].
// The SEV certificate chain is a singleton, so it uses a fixed key "chain".
func (s *SevCertChain) GetID() *[][]string {
	return &[][]string{{"chain"}}
}

// ToProto implements [state.WorldStateManager].
func (s *SevCertChain) ToProto() proto.Message {
	return s.CertificateChain
}

// ToWorldState implements [state.WorldStateManager].
func (s *SevCertChain) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, SEV_CERT_CHAIN_PREFIX, s.GetID(), s)
}

var _ state.WorldStateManager = (*SevCertChain)(nil)
