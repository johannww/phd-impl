package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// CreditScenario manages credit lifecycle operations
type CreditScenario struct {
	executor  *workload.Executor
	collector metrics.MetricsCollector
}

// NewCreditScenario creates a new credit scenario
func NewCreditScenario(executor *workload.Executor) *CreditScenario {
	return &CreditScenario{
		executor:  executor,
		collector: executor.GetCollector(),
	}
}

// MintCreditsContinuous runs a continuous minting loop.
// On every tick it fires one SubmitAsync per property; each commit is awaited
// in its own goroutine so the ticker loop never blocks on ordering latency.
func (s *CreditScenario) MintCreditsContinuous(
	ctx context.Context,
	gw *gateway.ClientWrapper,
	interval time.Duration,
	nProps int,
	quantityPerMint int64) error {
	tick := 0
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	ownerID := gw.GetIdentityID()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			timestamp := time.Now().Format(utils.RFC3339WithMillis)

			for propIdx := 0; propIdx < nProps; propIdx++ {
				propID := uint64(propIdx + 1)
				propertyID, _ := json.Marshal([]string{ownerID, fmt.Sprintf("%d", propID)})
				txID := fmt.Sprintf("mint-tick%d-prop%d", tick, propID)

				// Capture start time and submit without blocking the loop.
				start := time.Now()
				_, commit, err := gw.SubmitAsync(
					"MintQuantityCreditsForProperty",
					string(propertyID),
					strconv.FormatInt(quantityPerMint, 10),
					timestamp,
				)
				if err != nil {
					fmt.Printf("MintCreditsContinuous submit error tick %d prop %d: %v\n", tick, propID, err)
					recordTransaction(s.collector, txID, "credit-mint-continuous", start, err)
					continue
				}

				// Wait for commit in a goroutine; records end-to-end latency when done.
				awaitAndRecord(s.collector, txID, "credit-mint-continuous", start, commit, err, nil)
			}

			tick++
		}
	}
}

// BurnCredits burns credits with multipliers
func (s *CreditScenario) BurnCredits(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		quantity := int64((i + 1) * 100)
		burnCreditIDBytes, _ := json.Marshal([]string{fmt.Sprintf("burn-credit-%d", i)})

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction(
				"BurnNominalQuantity",
				string(burnCreditIDBytes),
				strconv.FormatInt(quantity, 10),
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		s.collector.Record(&metrics.TransactionMetric{
			ID:       fmt.Sprintf("burn-credit-%d", i),
			Scenario: "credit-burn",
			Latency:  time.Since(start),
			Success:  err == nil,
			Error:    getErrorString(err),
		})
	}

	return nil
}

// QueryCreditState reads credit state for verification
func (s *CreditScenario) QueryCreditState(ctx context.Context, client *gateway.ClientWrapper, creditID string) error {
	txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
		_, err := c.EvaluateTransaction("QueryCredit", creditID)
		return "", err
	}

	start := time.Now()
	_, err := txFunc(ctx, client)
	s.collector.Record(&metrics.TransactionMetric{
		ID:       fmt.Sprintf("query-credit-%s", creditID),
		Scenario: "credit-query",
		Latency:  time.Since(start),
		Success:  err == nil,
		Error:    getErrorString(err),
	})
	return nil
}
