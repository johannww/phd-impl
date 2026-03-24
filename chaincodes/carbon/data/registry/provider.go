package registry

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/identities"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const REGISTRY_PROVIDER_PREFIX = "registryProvider"

type RegistryProvider struct {
	Name    string `json:"name"`    // e.g., "SICAR"
	BaseURL string `json:"baseUrl"` // e.g., "https://data.sicar.gov.br"
	RootCA  []byte `json:"rootCa"`  // PEM encoded certificate
}

var _ state.WorldStateManager = (*RegistryProvider)(nil)

func (rp *RegistryProvider) ToProto() proto.Message {
	return &pb.RegistryProvider{
		Name:    rp.Name,
		BaseUrl: rp.BaseURL,
		RootCa:  rp.RootCA,
	}
}

func (rp *RegistryProvider) FromProto(m proto.Message) error {
	pr, ok := m.(*pb.RegistryProvider)
	if !ok {
		return fmt.Errorf("unexpected proto message type for RegistryProvider")
	}
	rp.Name = pr.Name
	rp.BaseURL = pr.BaseUrl
	rp.RootCA = pr.RootCa
	return nil
}

func (rp *RegistryProvider) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	if len(keyAttributes) == 0 {
		return fmt.Errorf("key attributes are empty")
	}
	return state.GetStateWithCompositeKey(stub, REGISTRY_PROVIDER_PREFIX, keyAttributes, rp)
}

func (rp *RegistryProvider) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, REGISTRY_PROVIDER_PREFIX, rp.GetID(), rp)
}

func (rp *RegistryProvider) GetID() *[][]string {
	return &[][]string{{rp.Name}}
}

func AddTrustedProvider(
	stub shim.ChaincodeStubInterface,
	name string,
	baseURL string,
	rootCA []byte,
) error {
	if err := cid.AssertAttributeValue(stub, identities.TrustedDBRegistrator, "true"); err != nil {
		return fmt.Errorf("caller is not a trusted database registrator: %v", err)
	}

	provider := &RegistryProvider{
		Name:    name,
		BaseURL: baseURL,
		RootCA:  rootCA,
	}

	return provider.ToWorldState(stub)
}
