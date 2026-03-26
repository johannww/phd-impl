package workload

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
)

// TransactionFunc is a function that executes a single transaction
type TransactionFunc func(ctx context.Context, client *gateway.ClientWrapper) (string, error)

// Executor orchestrates transaction execution with metrics collection
type Executor struct {
	client    *gateway.ClientWrapper
	collector metrics.MetricsCollector
	config    *ExecutorConfig
	limiter   *rate.Limiter
}

// ExecutorConfig holds executor configuration
type ExecutorConfig struct {
	ConcurrencyLevel  int           // Number of concurrent transactions
	TotalTransactions int           // Total transactions to execute
	Duration          time.Duration // If set, ignore TotalTransactions
	MetricsInterval   time.Duration // How often to print metrics
	TPS               float64       // Target throughput (transactions per second); 0 = unlimited
	BurstSize         int           // Token bucket burst size for rate limiting
}

// DefaultExecutorConfig returns default configuration
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		ConcurrencyLevel:  5,
		TotalTransactions: 100,
		MetricsInterval:   5 * time.Second,
	}
}

// NewExecutor creates a new transaction executor
func NewExecutor(client *gateway.ClientWrapper, cfg *ExecutorConfig) *Executor {
	if cfg == nil {
		cfg = DefaultExecutorConfig()
	}

	// Create rate limiter if TPS is configured
	var limiter *rate.Limiter
	if cfg.TPS > 0 {
		burst := cfg.BurstSize
		if burst <= 0 {
			burst = 1
		}
		limiter = rate.NewLimiter(rate.Limit(cfg.TPS), burst)
	}

	return &Executor{
		client:    client,
		collector: metrics.NewCollector(),
		config:    cfg,
		limiter:   limiter,
	}
}

// Execute runs transactions according to configuration
func (e *Executor) Execute(ctx context.Context, txFunc TransactionFunc, scenario string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create work queue
	workChan := make(chan int, e.config.ConcurrencyLevel)
	resultChan := make(chan error, e.config.ConcurrencyLevel)

	// Start worker goroutines
	for i := 0; i < e.config.ConcurrencyLevel; i++ {
		go e.worker(ctx, i, scenario, txFunc, workChan, resultChan)
	}

	// Track execution
	ticker := time.NewTicker(e.config.MetricsInterval)
	defer ticker.Stop()

	// Determine if we're running by duration or count
	var endTime time.Time
	var txCount int

	if e.config.Duration > 0 {
		endTime = time.Now().Add(e.config.Duration)
	} else {
		txCount = e.config.TotalTransactions
	}

	// Submit work
	go func() {
		txID := 0
		for {
			// Check if we should continue
			if e.config.Duration == 0 && txID >= txCount {
				break
			}
			if e.config.Duration > 0 && time.Now().After(endTime) {
				break
			}

			// Apply rate limiting if configured
			if e.limiter != nil {
				if err := e.limiter.Wait(ctx); err != nil {
					return // context cancelled
				}
			}

			select {
			case workChan <- txID:
				txID++
			case <-ctx.Done():
				return
			}
		}
		close(workChan)
	}()

	// Track completion
	completedTxs := 0
	expectedTxs := txCount
	if e.config.Duration > 0 {
		expectedTxs = -1 // Unknown
	}

	// Process results and print metrics
	for {
		select {
		case <-ticker.C:
			e.printMetrics(expectedTxs)

		case _, ok := <-resultChan:
			if !ok {
				return nil
			}
			completedTxs++

			if expectedTxs > 0 && completedTxs >= expectedTxs {
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// worker processes transactions from the work queue
func (e *Executor) worker(ctx context.Context, id int, scenario string, txFunc TransactionFunc, workChan <-chan int, resultChan chan<- error) {
	for txNum := range workChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		txID := fmt.Sprintf("%s-tx-%d-%d", scenario, id, txNum)
		start := time.Now()

		_, err := txFunc(ctx, e.client)
		latency := time.Since(start)

		metric := &metrics.TransactionMetric{
			ID:       txID,
			Scenario: scenario,
			Latency:  latency,
			Success:  err == nil,
			Error:    "",
		}

		if err != nil {
			metric.Error = err.Error()
		}

		e.collector.Record(metric)
		resultChan <- err
	}
}

// printMetrics prints current metrics
func (e *Executor) printMetrics(expectedTxs int) {
	reporter := metrics.NewReporter(e.collector)
	reporter.PrintLiveMetrics()
}

// GetCollector returns the metrics collector
func (e *Executor) GetCollector() metrics.MetricsCollector {
	return e.collector
}

// GetFinalReport returns the final metrics snapshot
func (e *Executor) GetFinalReport() *metrics.Snapshot {
	return e.collector.GetSnapshot()
}
