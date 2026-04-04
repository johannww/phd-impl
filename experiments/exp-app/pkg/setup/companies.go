package setup

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/johannww/phd-impl/chaincodes/carbon/companies"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
)

// SetupCompanies initializes one company metadata for the current identity
// and creates the pseudonym-to-companyID mapping required for auction participation.
func (s *SetupManager) SetupCompanies(ctx context.Context) ([]*client.Commit, error) {
	id := s.client.GetIdentityID()
	log.Printf("Preparing company for current identity: %s", id)

	company := &companies.Company{
		ID: id,
		Coordinate: &utils.Coordinate{
			Latitude:  -23.5505,
			Longitude: -46.6333,
		},
		DataProps: &data.ValidationProps{
			Methods: []data.ValidationMethod{data.ValidationMethodSattelite},
		},
	}

	compJSON, _ := json.Marshal(company)
	_, commit1, err := s.client.SubmitAsync("RegisterCompany", string(compJSON))
	if err != nil {
		log.Printf("Warning: Failed to submit register company %s: %v", id, err)
		return nil, nil
	}

	// Create pseudonym-to-companyID mapping (stores caller's pseudonym -> company.ID in private data)
	identityMapping := &companies.PseudonymToCompanyID{
		Pseudonym: id,
		CompanyID: company.ID,
	}
	mappingJSON, _ := json.Marshal(identityMapping)
	transient := map[string][]byte{
		companies.PSEUDONYM_TO_COMPANY_ID_TRANSIENT_KEY: mappingJSON,
	}

	_, commit2, err := s.client.SubmitAsyncWithTransient("CreatePseudonymToCompanyID", transient, "")
	if err != nil {
		log.Printf("Warning: Failed to create pseudonym mapping for %s: %v", id, err)
		return []*client.Commit{commit1}, nil
	}

	return []*client.Commit{commit1, commit2}, nil
}
