package setup

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/tee"
)

// InitializeBETS performs all necessary setup steps and returns a TEE client if enabled
func (s *SetupManager) InitializeBETS(ctx context.Context, nPropsPerOrg int, nChunksPerProp int) (*tee.Client, error) {
	log.Println("Starting BETS initialization...")

	var allCommits []*client.Commit

	// Call init function to initialize chaincode state
	_, commit, err := s.client.SubmitAsync("Init", "")
	allCommits = append(allCommits, commit)

	// 0. Setup SICAR
	commits, err := s.SetupSICAR(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SICAR: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 1. Set active policies
	commits, err = s.SetupPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup policies: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 2. Register Companies for each org
	commits, err = s.SetupCompanies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup companies: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 3. Register Buyer Wallets for each org
	commits, err = s.SetupBuyerWallets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup buyer wallets: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 4. Register Properties for each org
	commits, sicarData, err := s.SetupProperties(ctx, nPropsPerOrg, nChunksPerProp)
	if err != nil {
		return nil, fmt.Errorf("failed to setup properties: %v", err)
	}
	allCommits = append(allCommits, commits...)

	log.Printf("Waiting for %d setup transactions to commit...", len(allCommits))
	for i, commit := range allCommits {
		if _, err := commit.Status(); err != nil {
			log.Printf("Warning: Setup transaction commitment %d failed: %v", i+1, err)
		}
	}

	// 5. Refresh Registry Data for each property
	// This MUST be done after property registration is committed
	log.Println("Refreshing property registry data from SICAR...")
	refreshCommits, err := s.RefreshProperties(ctx, sicarData.SicarIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh properties: %v", err)
	}

	log.Printf("Waiting for %d refresh transactions to commit...", len(refreshCommits))
	for i, commit := range refreshCommits {
		if _, err := commit.Status(); err != nil {
			log.Printf("Warning: Refresh transaction commitment %d failed: %v", i+1, err)
		}
	}

	// 6. Setup TEE if enabled
	var teeClient *tee.Client
	if s.profile.TEEAuction.Enabled {
		log.Println("Setting up TEE auction service...")
		teeSetupMgr := NewTEESetupManager(s.client, s.profile, s.armTemplatePath)
		teeClient, err = teeSetupMgr.SetupTEE(ctx)
		if err != nil {
			log.Fatalf("WARNING: TEE setup failed: %v", err)
			log.Println("Continuing without TEE - will fall back to mock results")
			teeClient = nil
		} else {
			log.Println("TEE setup completed successfully")
		}
	}

	log.Println("BETS initialization complete.")
	return teeClient, nil
}
