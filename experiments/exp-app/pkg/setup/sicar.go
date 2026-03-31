package setup

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// SetupSICAR registers the SICAR mock service as a trusted provider.
// This is required for the "RefreshRegistryDataForProperty" transaction.
func (s *SetupManager) SetupSICAR(ctx context.Context) ([]*client.Commit, error) {
	if !s.profile.SICAR.Enabled {
		log.Println("SICAR setup skipped (disabled in profile)")
		return nil, nil
	}

	log.Println("Setting up SICAR trusted provider...")

	providerName := "SICAR"
	baseURL := "https://" + s.profile.SICAR.Endpoint

	// Load root CA from profile path
	rootCAPEM, err := os.ReadFile(s.profile.SICAR.Certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to read SICAR root CA from %s: %v", s.profile.SICAR.Certificate, err)
	}

	_, commit, err := s.client.SubmitAsync("AddTrustedProvider",
		providerName, baseURL, string(rootCAPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit AddTrustedProvider: %v", err)
	}

	return []*client.Commit{commit}, nil
}
