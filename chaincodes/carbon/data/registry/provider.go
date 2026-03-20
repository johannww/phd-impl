package registry

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/state"
)

const REGISTRY_PROVIDER_PREFIX = "registryProvider"

type RegistryProvider struct {
	Name    string `json:"name"`    // e.g., "SICAR"
	BaseURL string `json:"baseUrl"` // e.g., "https://data.sicar.gov.br"
	RootCA  []byte `json:"rootCa"`  // PEM encoded certificate
}

var _ state.WorldStateManager = (*RegistryProvider)(nil)

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
