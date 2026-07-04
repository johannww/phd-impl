package setup

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/tee"
)

// IdentitySetupResult contains setup artifacts for one identity slot.
type IdentitySetupResult struct {
	PropertyIDs []uint64
	SicarIDs    []string
}

// RunGlobalSetup executes global initialization steps that must run once per test run.
func (s *SetupManager) RunGlobalSetup(ctx context.Context) (*tee.Client, error) {
	log.Println("Starting global BETS setup...")

	var globalCommits []*client.Commit

	_, commit, err := s.client.SubmitAsync("Init", "")
	if err != nil {
		if isAlreadySetupError(err) {
			log.Printf("Init already applied, continuing: %v", err)
		} else {
			return nil, fmt.Errorf("failed to submit Init: %v", err)
		}
	} else {
		globalCommits = append(globalCommits, commit)
	}

	commits, err := s.SetupSICAR(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SICAR: %v", err)
	}
	globalCommits = append(globalCommits, commits...)

	commits, err = s.SetupPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup policies: %v", err)
	}
	globalCommits = append(globalCommits, commits...)

	log.Printf("Waiting for %d global setup transactions to commit...", len(globalCommits))
	for i, globalCommit := range globalCommits {
		if _, err := globalCommit.Status(); err != nil {
			log.Printf("Warning: Global setup transaction commitment %d failed: %v", i+1, err)
		}
	}

	if !s.profile.TEEAuction.Enabled {
		return nil, nil
	}

	log.Println("Setting up TEE auction service...")
	teeSetupMgr := NewTEESetupManager(s.client, s.profile, s.armTemplatePath)
	teeClient, err := teeSetupMgr.SetupTEE(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup TEE auction service: %v", err)
	}

	log.Println("Global BETS setup completed")
	return teeClient, nil
}

// InitializeBETS performs identity-scoped setup steps.
func (s *SetupManager) InitializeBETS(
	ctx context.Context,
	nPropsPerIdentity int,
	nChunksPerProp int,
	usersPerOrg int,
	assignment IdentityAssignment,
) (*IdentitySetupResult, error) {
	log.Printf(
		"Starting BETS initialization for org=%s userIndex=%d...",
		assignment.Organization,
		assignment.UserIndex,
	)

	var allCommits []*client.Commit
	var err error

	// 2. Register company for this identity slot
	commits, err := s.SetupCompanies(ctx, usersPerOrg, assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to setup companies: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 3. Register buyer wallet for this identity slot
	commits, err = s.SetupBuyerWallets(ctx, usersPerOrg, assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to setup buyer wallets: %v", err)
	}
	allCommits = append(allCommits, commits...)

	// 4. Register properties for this identity slot
	commits, sicarData, err := s.SetupProperties(ctx, nPropsPerIdentity, nChunksPerProp, usersPerOrg, assignment)
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

	if len(sicarData.PropertyIDs) == 0 {
		return nil, fmt.Errorf(
			"no properties were registered for org=%s userIndex=%d; cannot continue",
			assignment.Organization,
			assignment.UserIndex,
		)
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

	log.Printf(
		"BETS initialization complete for org=%s userIndex=%d with properties=%v",
		assignment.Organization,
		assignment.UserIndex,
		sicarData.PropertyIDs,
	)
	return &IdentitySetupResult{PropertyIDs: sicarData.PropertyIDs, SicarIDs: sicarData.SicarIDs}, nil
}
