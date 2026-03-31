package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	v "github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
)

// SetupProperties initializes nProps properties for each organization in the registry
func (s *SetupManager) SetupProperties(ctx context.Context, nPropsPerOrg int, nChunksPerProp int) ([]*client.Commit, error) {
	log.Printf("Preparing %d properties for each of the %d organizations...", nPropsPerOrg, len(s.profile.Peers))
	ownerID := s.client.GetIdentityID()

	var commits []*client.Commit
	propCounter := 1
	for orgName := range s.profile.Peers {
		log.Printf("Setting up %d properties for org %s (owned by current user %s)...", nPropsPerOrg, orgName, ownerID)
		for i := 0; i < nPropsPerOrg; i++ {
			propID := uint64(propCounter)
			propCounter++

			var chunks []*properties.PropertyChunk
			for j := 0; j < nChunksPerProp; j++ {
				chunk := &properties.PropertyChunk{
					PropertyID: propID,
					Coordinates: []utils.Coordinate{
						{
							Latitude:  -23.5505 + float64(propCounter)*0.01 + float64(j)*0.001,
							Longitude: -46.6333 + float64(propCounter)*0.01 + float64(j)*0.001,
						},
					},
					VegetationsProps: &v.VegetationProps{
						ForestType:    v.AtlanticForest,
						ForestDensity: 1.0,
						CropType:      v.Corn,
					},
					ValidationProps: &data.ValidationProps{
						Methods: []data.ValidationMethod{data.ValidationMethodSattelite},
					},
				}
				chunks = append(chunks, chunk)
			}

			property := &properties.Property{
				OwnerID:          ownerID,
				ID:               propID,
				RegistryID:       fmt.Sprintf("REG-%s-%d", orgName, propID),
				RegistryProvider: "SICAR",
				Chunks:           chunks,
			}

			propJSON, _ := json.Marshal(property)
			_, commit, err := s.client.SubmitAsync("RegisterProperty", string(propJSON))
			if err != nil {
				log.Printf("Warning: Failed to submit property %d for org %s: %v", propID, orgName, err)
				continue
			}
			commits = append(commits, commit)
		}
	}

	return commits, nil
}
