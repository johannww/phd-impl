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

// BiddingScenario creates buy and sell bids for auctions
type BiddingScenario struct {
	executor  *workload.Executor
	collector metrics.MetricsCollector
}

// NewBiddingScenario creates a new bidding scenario
func NewBiddingScenario(executor *workload.Executor) *BiddingScenario {
	return &BiddingScenario{
		executor:  executor,
		collector: executor.GetCollector(),
	}
}

// CreateBuyBids creates N buy bids with varying quantities
func (s *BiddingScenario) CreateBuyBids(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Create buy bid with varying quantity
		quantity := int64((i + 1) * 100) // 100, 200, 300...

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction("CreateBuyBidPublicQuantity", strconv.FormatInt(quantity, 10))
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("buy-bid-%d", i),
			Scenario: "bidding-buy",
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

// CreateSellBids creates N sell bids with varying prices
func (s *BiddingScenario) CreateSellBids(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Create sell bid with varying quantity
		quantity := int64((i + 1) * 50) // 50, 100, 150...
		// Example credit ID - in real scenario this would come from minted credits
		creditID := fmt.Sprintf("credit-%d", i)

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction(
				"CreateSellBidFromWallet",
				strconv.FormatInt(quantity, 10),
				creditID,
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("sell-bid-%d", i),
			Scenario: "bidding-sell",
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

// ReadBids reads existing bids for verification
func (s *BiddingScenario) ReadBids(ctx context.Context, client *gateway.ClientWrapper, bidID string) error {
	txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
		// This would be a read-only query function once implemented
		_, err := c.EvaluateTransaction("QueryBid", bidID)
		return "", err
	}

	start := time.Now()
	_, err := txFunc(ctx, client)
	latency := time.Since(start)

	metric := &metrics.TransactionMetric{
		ID:       fmt.Sprintf("read-bid-%s", bidID),
		Scenario: "bidding-read",
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
