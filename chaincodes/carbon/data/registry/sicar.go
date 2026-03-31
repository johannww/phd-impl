package registry

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

// SicarData represents the specific structure returned by the SICAR mock API.
type SicarData struct {
	CodigoImovel                    string  `json:"codigoImovel"`
	SituacaoImovel                  string  `json:"situacaoImovel"`
	AreaTotalImovel                 float64 `json:"areaTotalImovel"`
	AreaPreservacaoPermanente       float64 `json:"areaPreservacaoPermanente"`
	AreaReservaLegalDeclarada       float64 `json:"areaReservaLegalDeclaradaProprietarioPossuidor"`
	AreaRemanescenteVegetacaoNativa float64 `json:"areaRemanescenteVegetacaoNativa"`
}

// ToSummary converts SICAR-specific data into a generic RegistrySummary.
func (s *SicarData) ToSummary() *RegistrySummary {
	summary := &RegistrySummary{
		RegistryPropID:  s.CodigoImovel,
		Status:          s.SituacaoImovel,
		TotalArea:       s.AreaTotalImovel,
		LegalForestArea: s.AreaPreservacaoPermanente + s.AreaReservaLegalDeclarada,
		VerifiedForest:  s.AreaRemanescenteVegetacaoNativa,
	}

	summary.Status = INACTIVE_STATUS
	if s.SituacaoImovel == "AT" || s.SituacaoImovel == "RE" {
		summary.Status = ACTIVE_STATUS
	}
	return summary
}

// RefreshRegistryData fetches, verifies, and saves SICAR data for a given registry ID to the world state.
func RefreshRegistryData(
	stub shim.ChaincodeStubInterface,
	providerName string,
	registryPropID string,
) (*RegistrySummary, error) {
	// For SICAR, we use the demonstrativoDegustacao endpoint
	endpoint := fmt.Sprintf("/sicar/demonstrativoDegustacao/1.0/%s", registryPropID)

	bytes, err := FetchVerifiedData(stub, providerName, endpoint)
	if err != nil {
		return nil, err
	}

	// The SICAR mock returns a "result" wrapper (as seen in data_api/internal/sicar/types.go)
	var response struct {
		Result []SicarData `json:"result"`
	}

	if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SICAR response: %v", err)
	}

	if len(response.Result) == 0 {
		return nil, fmt.Errorf("no SICAR data found for ID %s", registryPropID)
	}

	summary := response.Result[0].ToSummary()
	if err := summary.ToWorldState(stub); err != nil {
		return nil, fmt.Errorf("failed to save registry summary to world state: %v", err)
	}

	return summary, nil
}
