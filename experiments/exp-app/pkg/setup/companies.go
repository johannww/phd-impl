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
	_, commit, err := s.client.SubmitAsync("RegisterCompany", string(compJSON))
	if err != nil {
		log.Printf("Warning: Failed to submit register company %s: %v", id, err)
		return nil, nil
	}

	return []*client.Commit{commit}, nil
}
