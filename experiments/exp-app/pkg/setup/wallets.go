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

	commits := make([]*client.Commit, 0, s.walletsPerBuyer)
	for walletNumber := int64(0); walletNumber < int64(s.walletsPerBuyer); walletNumber++ {
		_, commit, submitErr := s.client.SubmitAsync(
			"MintVirtualTokenForWalletId",
			id,
			strconv.FormatInt(walletNumber, 10),
			strconv.FormatInt(initialBalance, 10),
		)
		if submitErr != nil {
			log.Printf("Warning: Failed to submit wallet setup for %s wallet %d: %v", id, walletNumber, submitErr)
			return nil, fmt.Errorf("failed to submit wallet setup for %s wallet %d: %v", id, walletNumber, submitErr)
		}
		commits = append(commits, commit)
	}

	return commits, nil
}
