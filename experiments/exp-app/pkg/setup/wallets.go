package setup

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// SetupBuyerWallets initializes one virtual token wallet for the current identity.
func (s *SetupManager) SetupBuyerWallets(
	ctx context.Context,
	usersPerOrg int,
	assignment IdentityAssignment,
) ([]*client.Commit, error) {
	_, _, err := s.resolveIdentitySlot(usersPerOrg, assignment)
	if err != nil {
		return nil, err
	}

	id := s.client.GetIdentityID()
	log.Printf(
		"Preparing wallet for current identity: %s (org=%s userIndex=%d)",
		id,
		assignment.Organization,
		assignment.UserIndex,
	)
	initialBalance := int64(math.MaxInt64 >> 1) // Use a very large initial balance for testing

	callerID, err := s.client.SubmitTransaction("ReturnCallerID")
	if err != nil {
		log.Printf("Warning: Failed to get caller ID for wallet setup: %v", err)
		return nil, fmt.Errorf("failed to get caller ID for wallet setup: %v", err)
	}

	if string(callerID) != id {
		return nil, fmt.Errorf("caller ID mismatch during wallet setup. expected %s, got %s", id, string(callerID))
	}

	_, commit, err := s.client.SubmitAsync(
		"MintVirtualToken",
		id,
		strconv.FormatInt(initialBalance, 10),
	)
	if err != nil {
		log.Printf("Warning: Failed to submit wallet setup for %s: %v", id, err)
		return nil, fmt.Errorf("failed to submit wallet setup for %s: %v", id, err)
	}

	return []*client.Commit{commit}, nil
}
