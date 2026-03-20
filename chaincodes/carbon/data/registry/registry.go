package registry

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/state"
)

const REGISTRY_SUMMARY_PREFIX = "registrySummary"

// RegistrySummary represents a generic environmental registry response.
// This is the canonical format used by the chaincode regardless of the data source.
type RegistrySummary struct {
	RegistryPropID  string  `json:"registryPropId"`  // e.g., SICAR property ID
	Status          string  `json:"status"`          // e.g., "ACTIVE", "PENDING"
	TotalArea       float64 `json:"totalArea"`       // in hectares
	LegalForestArea float64 `json:"legalForestArea"` // APP + Legal Reserve
	VerifiedForest  float64 `json:"verifiedForest"`  // Actual forest area
}

var _ state.WorldStateManager = (*RegistrySummary)(nil)

func (r *RegistrySummary) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, REGISTRY_SUMMARY_PREFIX, keyAttributes, r)
}

func (r *RegistrySummary) ToWorldState(stub shim.ChaincodeStubInterface) error {
	return state.PutStateWithCompositeKey(stub, REGISTRY_SUMMARY_PREFIX, r.GetID(), r)
}

func (r *RegistrySummary) GetID() *[][]string {
	return &[][]string{{r.RegistryPropID}}
}
