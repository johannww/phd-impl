package setup

import (
	"context"
	"log"
	"math"
	"strconv"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// SetupBuyerWallets initializes one virtual token wallet for the current identity
func (s *SetupManager) SetupBuyerWallets(ctx context.Context) ([]*client.Commit, error) {
	id := s.client.GetIdentityID()
	log.Printf("Preparing wallet for current identity: %s", id)
	initialBalance := int64(math.MaxInt64 >> 1) // Use a very large initial balance for testing

	callerID, err := s.client.SubmitTransaction("ReturnCallerID")
	if err != nil {
		log.Printf("Warning: Failed to get caller ID for wallet setup: %v", err)
		return nil, nil
	}

	if string(callerID) != id {
		log.Fatalf("Warning: Caller ID mismatch during wallet setup. Expected %s, got %s", id, string(callerID))
		return nil, nil
	}

	_, commit, err := s.client.SubmitAsync(
		"MintVirtualToken",
		id,
		strconv.FormatInt(initialBalance, 10),
	)
	if err != nil {
		log.Printf("Warning: Failed to submit wallet setup for %s: %v", id, err)
		return nil, nil
	}

	return []*client.Commit{commit}, nil
}
