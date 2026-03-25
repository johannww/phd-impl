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

// CreditScenario manages credit lifecycle operations
type CreditScenario struct {
	executor  *workload.Executor
	collector *metrics.Collector
}

// NewCreditScenario creates a new credit scenario
func NewCreditScenario(executor *workload.Executor) *CreditScenario {
	return &CreditScenario{
		executor:  executor,
		collector: executor.GetCollector(),
	}
}

// MintCredits creates credits for properties
func (s *CreditScenario) MintCredits(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		quantity := int64((i + 1) * 1000) // 1000, 2000, 3000...
		propertyID := fmt.Sprintf("property-%d", i)

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction(
				"MintQuantityCreditForChunk",
				propertyID,
				strconv.FormatInt(quantity, 10),
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("mint-credit-%d", i),
			Scenario: "credit-mint",
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

// TransferCredits transfers credits from mint to user wallet
func (s *CreditScenario) TransferCredits(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		quantity := int64((i + 1) * 500)
		mintCreditID := fmt.Sprintf("mint-credit-%d", i)

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			// TransferFromMintToWallet expects composite key as array
			_, err := c.SubmitTransaction(
				"TransferFromMintToWallet",
				mintCreditID,
				strconv.FormatInt(quantity, 10),
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("transfer-credit-%d", i),
			Scenario: "credit-transfer",
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

// BurnCredits burns credits with multipliers
func (s *CreditScenario) BurnCredits(ctx context.Context, client *gateway.ClientWrapper, count int) error {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		quantity := int64((i + 1) * 100)
		burnCreditID := fmt.Sprintf("burn-credit-%d", i)

		txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
			_, err := c.SubmitTransaction(
				"BurnNominalQuantity",
				burnCreditID,
				strconv.FormatInt(quantity, 10),
			)
			return "", err
		}

		start := time.Now()
		_, err := txFunc(ctx, client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       fmt.Sprintf("burn-credit-%d", i),
			Scenario: "credit-burn",
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

// QueryCreditState reads credit state for verification
func (s *CreditScenario) QueryCreditState(ctx context.Context, client *gateway.ClientWrapper, creditID string) error {
	txFunc := func(ctx context.Context, c *gateway.ClientWrapper) (string, error) {
		_, err := c.EvaluateTransaction("QueryCredit", creditID)
		return "", err
	}

	start := time.Now()
	_, err := txFunc(ctx, client)
	latency := time.Since(start)

	metric := &metrics.TransactionMetric{
		ID:       fmt.Sprintf("query-credit-%s", creditID),
		Scenario: "credit-query",
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
