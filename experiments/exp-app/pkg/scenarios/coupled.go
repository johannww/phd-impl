package scenarios

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// CoupledAuctionScenario manages coupled auction operations
type CoupledAuctionScenario struct {
	executor  *workload.Executor
	collector *metrics.Collector
}

// NewCoupledAuctionScenario creates a new coupled auction scenario
func NewCoupledAuctionScenario(executor *workload.Executor) *CoupledAuctionScenario {
	return &CoupledAuctionScenario{
		executor:  executor,
		collector: executor.GetCollector(),
	}
}

// SetAuctionType sets the auction type to coupled
func (s *CoupledAuctionScenario) SetAuctionType(ctx context.Context, client *gateway.ClientWrapper) error {
	txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
		_, err := c.SubmitTransaction("SetAuctionType", "coupled")
		return "", err
	}

	start := time.Now()
	_, err := txFunc(ctx, client)
	latency := time.Since(start)

	metric := &metrics.TransactionMetric{
		ID:       "set-auction-type-coupled",
		Scenario: "coupled-auction-setup",
		Latency:  latency,
		Success:  err == nil,
		Error:    "",
	}

	if err != nil {
		metric.Error = err.Error()
	}

	s.collector.Record(metric)
	return err
}

// PublishOffChainResults publishes off-chain auction results
func (s *CoupledAuctionScenario) PublishOffChainResults(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		auctionID := fmt.Sprintf("auction-%d", i)
		// This would contain serialized auction result data in real scenario
		resultData := fmt.Sprintf(`{"auctionID":"%s","matchedBidsPublic":[]}`, auctionID)

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction(
				"ProcessOffChainAuctionResult",
				resultData,
				resultData, // public and private results
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("publish-result-%d", i),
			Scenario: "coupled-auction-publish",
			Latency:  latency,
			Success:  err == nil,
			Error:    "",
		}

		if err != nil {
			metric.Error = err.Error()
		}

		s.collector.Record(metric)
	}

	return nil
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

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction(
				"UpdateSellerAndBuyerVirtualTokenWallets",
				userID,
				strconv.FormatInt(quantity, 10),
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("update-wallet-%d", i),
			Scenario: "coupled-auction-wallet",
			Latency:  latency,
			Success:  err == nil,
			Error:    "",
		}

		if err != nil {
			metric.Error = err.Error()
		}

		s.collector.Record(metric)
	}

	return nil
}

// QueryAuctionResults queries auction results
func (s *CoupledAuctionScenario) QueryAuctionResults(ctx context.Context, client *gateway.ClientWrapper, auctionID string) error {
	txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
		_, err := c.EvaluateTransaction("QueryAuctionResults", auctionID)
		return "", err
	}

	start := time.Now()
	_, err := txFunc(ctx, client)
	latency := time.Since(start)

	metric := &metrics.TransactionMetric{
		ID:       fmt.Sprintf("query-auction-results-%s", auctionID),
		Scenario: "coupled-auction-query",
		Latency:  latency,
		Success:  err == nil,
		Error:    "",
	}

	if err != nil {
		metric.Error = err.Error()
	}

	s.collector.Record(metric)
	return nil
}
