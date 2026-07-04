package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/carbon/properties"
	v "github.com/johannww/phd-impl/chaincodes/carbon/vegetation"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
)

type SicarData struct {
	SicarIDs    []string
	PropertyIDs []uint64
}

// IdentityAssignment identifies which org/user slot this setup manager should initialize.
type IdentityAssignment struct {
	Organization string
	UserIndex    int // 0-based index within organization
}

// SetupProperties initializes nPropsPerIdentity properties for one deterministic identity slot.
// Property IDs and SICAR IDs are deterministically partitioned across org/user slots to avoid collisions.
func (s *SetupManager) SetupProperties(
	ctx context.Context,
	nPropsPerIdentity int,
	nChunksPerProp int,
	usersPerOrg int,
	assignment IdentityAssignment,
) ([]*client.Commit, *SicarData, error) {
	if nPropsPerIdentity <= 0 {
		return nil, nil, fmt.Errorf("nPropsPerIdentity must be > 0")
	}
	if usersPerOrg <= 0 {
		return nil, nil, fmt.Errorf("usersPerOrg must be > 0")
	}

	ownerID := s.client.GetIdentityID()
	log.Printf("Preparing %d properties for identity %s (org=%s userIndex=%d)", nPropsPerIdentity, ownerID, assignment.Organization, assignment.UserIndex)

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
	sort.Strings(sicarIDs)
	if len(sicarIDs) == 0 {
		return nil, nil, fmt.Errorf("no SICAR IDs found in sicar.json")
	}

	slot, totalIdentities, err := s.resolveIdentitySlot(usersPerOrg, assignment)
	if err != nil {
		return nil, nil, err
	}

	requiredSicar := totalIdentities * nPropsPerIdentity
	if len(sicarIDs) < requiredSicar {
		return nil, nil, fmt.Errorf(
			"insufficient SICAR properties: required=%d available=%d (totalIdentities=%d nPropsPerIdentity=%d)",
			requiredSicar,
			len(sicarIDs),
			totalIdentities,
			nPropsPerIdentity,
		)
	}

	start := slot * nPropsPerIdentity
	end := start + nPropsPerIdentity
	assignedSicar := sicarIDs[start:end]

	var commits []*client.Commit
	registeredSicarIDs := make([]string, 0, nPropsPerIdentity)
	registeredPropertyIDs := make([]uint64, 0, nPropsPerIdentity)
	for i := 0; i < nPropsPerIdentity; i++ {
		propID := uint64(start + i + 1)
		sicarID := assignedSicar[i]

		var chunks []*properties.PropertyChunk
		for j := 0; j < nChunksPerProp; j++ {
			chunk := &properties.PropertyChunk{
				PropertyID: propID,
				Coordinates: []utils.Coordinate{
					{
						Latitude:  -23.5505 + float64(start+i+1)*0.01 + float64(j)*0.001,
						Longitude: -46.6333 + float64(start+i+1)*0.01 + float64(j)*0.001,
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
		_, commit, submitErr := s.client.SubmitAsync("RegisterProperty", string(propJSON))
		if submitErr != nil {
			if isAlreadySetupError(submitErr) {
				log.Printf("Property %d already registered for owner=%s (sicar=%s), reusing", propID, ownerID, sicarID)
				registeredSicarIDs = append(registeredSicarIDs, sicarID)
				registeredPropertyIDs = append(registeredPropertyIDs, propID)
				continue
			}
			log.Printf("Warning: Failed to submit property %d (owner=%s sicar=%s): %v", propID, ownerID, sicarID, submitErr)
			continue
		}
		commits = append(commits, commit)
		registeredSicarIDs = append(registeredSicarIDs, sicarID)
		registeredPropertyIDs = append(registeredPropertyIDs, propID)
	}

	return commits, &SicarData{SicarIDs: registeredSicarIDs, PropertyIDs: registeredPropertyIDs}, nil
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
