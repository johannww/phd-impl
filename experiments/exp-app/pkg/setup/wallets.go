package setup

import (
	"context"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// SetupBuyerWallets initializes one virtual token wallet for the current identity
func (s *SetupManager) SetupBuyerWallets(ctx context.Context) ([]*client.Commit, error) {
	id := s.client.GetIdentityID()
	log.Printf("Preparing wallet for current identity: %s", id)
	initialBalance := int64(1000000000) // Large enough for many auctions

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
