package setup

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// InitializeBETS performs all necessary setup steps
func (s *SetupManager) InitializeBETS(ctx context.Context, nPropsPerOrg int, nChunksPerProp int) error {
	log.Println("Starting BETS initialization...")

	var allCommits []*client.Commit

	// 0. Setup SICAR
	commits, err := s.SetupSICAR(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup SICAR: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 1. Register Companies for each org
	commits, err = s.SetupCompanies(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup companies: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 2. Register Buyer Wallets for each org
	commits, err = s.SetupBuyerWallets(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup buyer wallets: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 3. Register Properties for each org
	commits, err = s.SetupProperties(ctx, nPropsPerOrg, nChunksPerProp)
	if err != nil {
		return fmt.Errorf("failed to setup properties: %v", err)
	}
	allCommits = append(allCommits, commits...)

	log.Printf("Waiting for %d setup transactions to commit...", len(allCommits))
	for i, commit := range allCommits {
		if _, err := commit.Status(); err != nil {
			log.Printf("Warning: Setup transaction commitment %d failed: %v", i+1, err)
		}
	}

	log.Println("BETS initialization complete.")
	return nil
}
