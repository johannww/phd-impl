package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// SetupPolicies sets the active mint/burn policies on the chaincode.
func (s *SetupManager) SetupPolicies(ctx context.Context) ([]*client.Commit, error) {
	log.Println("Setting active policies...")

	activePolicies := []string{"vegetation", "registry_freshness"}
	policiesJSON, err := json.Marshal(activePolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal active policies: %v", err)
	}

	_, commit, err := s.client.SubmitAsync("SetActivePolicies", string(policiesJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to submit SetActivePolicies: %v", err)
	}

	return []*client.Commit{commit}, nil
}
