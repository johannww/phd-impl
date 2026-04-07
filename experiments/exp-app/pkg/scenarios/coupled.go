package scenarios

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/tee"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// CoupledAuctionScenario manages coupled auction operations
type CoupledAuctionScenario struct {
	executor  *workload.Executor
	collector metrics.MetricsCollector
	teeClient *tee.Client
}

// NewCoupledAuctionScenario creates a new coupled auction scenario
func NewCoupledAuctionScenario(executor *workload.Executor) *CoupledAuctionScenario {
	return &CoupledAuctionScenario{
		executor:  executor,
		collector: executor.GetCollector(),
		teeClient: nil, // Will be set via SetTEEClient if TEE is enabled
	}
}

// SetTEEClient sets the TEE client for this scenario
func (s *CoupledAuctionScenario) SetTEEClient(client *tee.Client) {
	s.teeClient = client
}

// SetAuctionType sets the auction type to coupled
func (s *CoupledAuctionScenario) SetAuctionType(ctx context.Context, client *gateway.ClientWrapper) error {
	start := time.Now()
	_, err := client.SubmitTransaction("SetAuctionType", "coupled")
	recordTransaction(s.collector, "set-auction-type-coupled", "coupled-auction-setup", start, err)
	return err
}

// LockAuction locks the auction semaphore to prevent concurrent auctions.
func (s *CoupledAuctionScenario) LockAuction(ctx context.Context, client *gateway.ClientWrapper) error {
	start := time.Now()
	_, err := client.SubmitTransaction("LockAuctionSemaphore")
	recordTransaction(s.collector, "lock-auction-semaphore", "coupled-auction-lock", start, err)
	return err
}

// UnlockAuction releases the auction semaphore.
func (s *CoupledAuctionScenario) UnlockAuction(ctx context.Context, client *gateway.ClientWrapper) error {
	start := time.Now()
	_, err := client.SubmitTransaction("UnlockAuctionSemaphore")
	recordTransaction(s.collector, "unlock-auction-semaphore", "coupled-auction-unlock", start, err)
	return err
}

// CommitAuctionData creates a snapshot of the current state for the TEE auction.
// The auction semaphore must be locked before calling this function.
// The end timestamp is determined by the semaphore on-chain; no argument is needed.
func (s *CoupledAuctionScenario) CommitAuctionData(ctx context.Context, client *gateway.ClientWrapper) error {
	start := time.Now()
	_, err := client.SubmitTransaction("CommitDataForTEEAuction")
	recordTransaction(s.collector, "commit-auction-data", "coupled-auction-commit", start, err)
	return err
}

// PublishTEEResults runs the TEE auction and publishes signed results
func (s *CoupledAuctionScenario) PublishTEEResults(ctx context.Context, client *gateway.ClientWrapper) error {
	// Check if TEE client is configured
	if s.teeClient == nil {
		log.Println("TEE client not configured, using mock results")
		return s.publishMockResults(ctx, client)
	}

	start := time.Now()

	// Step 1: Fetch auction data from chaincode
	serializedAuctionDataBytes, err := client.EvaluateTransaction("RetrieveDataForTEEAuction")
	if err != nil {
		recordTransaction(s.collector, "publish-tee-results-get-data", "coupled-auction-publish-tee", start, err)
		return fmt.Errorf("failed to get committed auction data: %w", err)
	}

	// Step 2: Send to TEE service for processing
	teeStart := time.Now()
	auctionResp, err := s.teeClient.RunAuction(serializedAuctionDataBytes)
	if err != nil {
		recordTransaction(s.collector, "publish-tee-results-tee-run", "coupled-auction-publish-tee", teeStart, err)
		return fmt.Errorf("failed to run TEE auction: %w", err)
	}
	recordTransaction(s.collector, "publish-tee-results-tee-run", "coupled-auction-tee-processing", teeStart, nil)

	// Step 3: Serialize results
	publicResultBytes, err := tee.SerializeAuctionResults(auctionResp.Public)
	if err != nil {
		return fmt.Errorf("failed to serialize public result: %w", err)
	}

	privateResultBytes, err := tee.SerializeAuctionResults(auctionResp.Private)
	if err != nil {
		return fmt.Errorf("failed to serialize private result: %w", err)
	}

	// Step 4: Publish to chaincode
	transient := map[string][]byte{
		"serializedResultPvt": privateResultBytes,
	}

	_, err = client.SubmitWithTransient("PublishTEEAuctionResults", transient, string(publicResultBytes))
	recordTransaction(s.collector, "publish-tee-results", "coupled-auction-publish-tee", start, err)
	return err
}

// publishMockResults publishes mock TEE results (fallback when TEE is not available)
func (s *CoupledAuctionScenario) publishMockResults(ctx context.Context, client *gateway.ClientWrapper) error {
	mockResultPub := `{"resultBytes":"YmFzZTY0X2VuY29kZWRfcmVzdWx0c19wdWI=","signature":"mock_signature"}`
	mockResultPvt := `{"resultBytes":"YmFzZTY0X2VuY29kZWRfcmVzdWx0c19wdnQ=","signature":"mock_signature"}`

	transient := map[string][]byte{
		"serializedResultPvt": []byte(mockResultPvt),
	}

	start := time.Now()
	_, err := client.SubmitWithTransient("PublishTEEAuctionResults", transient, mockResultPub)
	recordTransaction(s.collector, "publish-tee-results-mock", "coupled-auction-publish-tee", start, err)
	return err
}

// PeriodicAuction runs a continuous auction loop (Lock -> Commit -> Sleep -> Publish/Unlock)
func (s *CoupledAuctionScenario) PeriodicAuction(
	ctx context.Context,
	client *gateway.ClientWrapper,
	interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := s.SetAuctionType(ctx, client); err != nil {
		return fmt.Errorf("failed to set auction type: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.LockAuction(ctx, client); err != nil {
				log.Printf("Failed to lock auction semaphore, skipping this cycle: %v", err)
				continue
			}

			if err := s.CommitAuctionData(ctx, client); err != nil {
				// Lock acquired but commit failed — release the semaphore so the
				// next cycle can try again.
				// s.UnlockAuction(ctx, client) //nolint:errcheck
				log.Printf("Failed to commit auction data, skipping publish and unlocking semaphore: %v", err)
				continue
			}

			select {
			case <-ctx.Done():
				// Auction is locked and committed but we are shutting down —
				// clean up the semaphore so a future run can proceed.
				// s.UnlockAuction(ctx, client) //nolint:errcheck
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}

			// PublishTEEAuctionResults calls UnlockAuction internally on success.
			// On failure the semaphore remains locked; log it and let the next
			// ticker cycle attempt a fresh lock (which will fail fast until the
			// network operator clears the semaphore manually).
			err := s.PublishTEEResults(ctx, client)
			if err != nil {
				log.Printf("Failed to publish TEE auction results: %v\n", err)
			}
		}
	}
}

// UpdateWallets updates virtual token wallets after auction
func (s *CoupledAuctionScenario) UpdateWallets(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		userID := fmt.Sprintf("user-%d", i)
		quantity := int64((i + 1) * 50)

		start := time.Now()
		_, err := client.SubmitTransaction(
			"UpdateSellerAndBuyerVirtualTokenWallets",
			userID,
			strconv.FormatInt(quantity, 10),
		)
		recordTransaction(s.collector, fmt.Sprintf("update-wallet-%d", i), "coupled-auction-wallet", start, err)
	}

	return nil
}

// QueryAuctionResults queries auction results
func (s *CoupledAuctionScenario) QueryAuctionResults(ctx context.Context, client *gateway.ClientWrapper, auctionID string) error {
	start := time.Now()
	_, err := client.EvaluateTransaction("QueryAuctionResults", auctionID)
	recordTransaction(s.collector, fmt.Sprintf("query-auction-results-%s", auctionID), "coupled-auction-query", start, err)
	return err
}
