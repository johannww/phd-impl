package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"math/rand"
	"os"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	v "github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
)

type SicarData struct {
	SicarIDs []string
}

// SetupProperties initializes nProps properties for each organization in the registry
func (s *SetupManager) SetupProperties(ctx context.Context, nPropsPerOrg int, nChunksPerProp int) ([]*client.Commit, *SicarData, error) {
	log.Printf("Preparing %d properties for each of the %d organizations...", nPropsPerOrg, len(s.profile.Peers))
	ownerID := s.client.GetIdentityID()

	// Load SICAR IDs from profile-defined path
	sicarFile, err := os.ReadFile(s.profile.SICAR.DataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read SICAR data from %s: %v", s.profile.SICAR.DataPath, err)
	}
	var sicarData map[string]interface{}
	if err := json.Unmarshal(sicarFile, &sicarData); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal sicar.json: %v", err)
	}
	var sicarIDs []string
	for id := range sicarData {
		sicarIDs = append(sicarIDs, id)
	}
	if len(sicarIDs) == 0 {
		return nil, nil, fmt.Errorf("no SICAR IDs found in sicar.json")
	}

	rand.Seed(time.Now().UnixNano())

	var commits []*client.Commit
	propCounter := 1
	for orgName := range s.profile.Peers {
		log.Printf("Setting up %d properties for org %s (owned by current user %s)...", nPropsPerOrg, orgName, ownerID)
		for i := 0; i < nPropsPerOrg; i++ {
			propID := uint64(propCounter)
			propCounter++

			// Select a random SICAR ID
			sicarID := sicarIDs[rand.Intn(len(sicarIDs))]

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
				RegistryID:       sicarID,
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

	return commits, &SicarData{SicarIDs: sicarIDs}, nil
}

// RefreshProperties updates the registry data for each property in the world state
func (s *SetupManager) RefreshProperties(ctx context.Context, sicarIDs []string) ([]*client.Commit, error) {
	log.Printf("Refreshing registry data for %d properties...", len(sicarIDs))

	var commits []*client.Commit
	for _, sicarID := range sicarIDs {
		// The chaincode expects providerName (SICAR) and registryPropID (sicarID)
		_, commit, err := s.client.SubmitAsync("RefreshRegistryDataForProperty", "SICAR", sicarID)
		if err != nil {
			log.Printf("Warning: Failed to submit refresh for property %s: %v", sicarID, err)
			continue
		}
		commits = append(commits, commit)
	}

	return commits, nil
}
