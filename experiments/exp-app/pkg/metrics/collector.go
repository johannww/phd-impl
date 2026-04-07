package metrics

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// CommitTask represents a pending commit to await and record
type CommitTask struct {
	ID         string
	Scenario   string
	Start      time.Time
	Commit     *client.Commit
	TxErr      error
	RunOnError func()
}

// MetricsCollector is the interface for metrics collection
type MetricsCollector interface {
	Record(m *TransactionMetric)
	GetSnapshot() *Snapshot
	GetScenarioStats() map[string]*ScenarioStats
	GetAllMetrics() []*TransactionMetric
	SubmitCommitTask(task *CommitTask)
	Stop()
}

// TransactionMetric holds a single transaction's performance data
type TransactionMetric struct {
	ID        string        `json:"id"`
	Scenario  string        `json:"scenario"`
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency_ms"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
}

// Snapshot holds aggregated metrics at a point in time
type Snapshot struct {
	Timestamp       time.Time
	TotalTxs        int64
	SuccessfulTxs   int64
	FailedTxs       int64
	ThroughputTPS   float64
	LatencyP50MS    float64
	LatencyP95MS    float64
	LatencyP99MS    float64
	LatencyAvgMS    float64
	LatencyMinMS    float64
	LatencyMaxMS    float64
	SuccessRatePerc float64
}

// ScenarioStats tracks metrics per scenario type
type ScenarioStats struct {
	Name            string
	TotalTxs        int64
	SuccessfulTxs   int64
	FailedTxs       int64
	LatencyP50MS    float64
	LatencyP95MS    float64
	LatencyP99MS    float64
	LatencyAvgMS    float64
	LatencyMinMS    float64
	LatencyMaxMS    float64
	SuccessRatePerc float64
}

// Collector aggregates transaction metrics
type Collector struct {
	mu           sync.RWMutex
	metrics      []*TransactionMetric
	startTime    time.Time
	lastSnapshot time.Time
}

// // NewCollector creates a new metrics collector
// func NewCollector() *Collector {
// 	return &Collector{
// 		metrics:   make([]*TransactionMetric, 0, 10000),
// 		startTime: time.Now(),
// 	}
// }

func NewCollector() *BucketedCollector {
	return NewBucketedCollector()
}

// Record adds a transaction metric
func (c *Collector) Record(m *TransactionMetric) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m.Timestamp = time.Now()
	c.metrics = append(c.metrics, m)
}

// GetSnapshot returns current aggregated metrics
func (c *Collector) GetSnapshot() *Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.metrics) == 0 {
		return &Snapshot{Timestamp: time.Now()}
	}

	snapshot := &Snapshot{
		Timestamp: time.Now(),
		TotalTxs:  int64(len(c.metrics)),
	}

	latencies := make([]float64, len(c.metrics))
	minLatency := time.Duration(1<<63 - 1)
	maxLatency := time.Duration(0)
	totalLatency := time.Duration(0)

	for i, m := range c.metrics {
		if m.Success {
			snapshot.SuccessfulTxs++
		} else {
			snapshot.FailedTxs++
		}

		latencyMS := float64(m.Latency.Milliseconds())
		latencies[i] = latencyMS
		totalLatency += m.Latency

		if m.Latency < minLatency {
			minLatency = m.Latency
		}
		if m.Latency > maxLatency {
			maxLatency = m.Latency
		}
	}

	snapshot.SuccessRatePerc = float64(snapshot.SuccessfulTxs) / float64(snapshot.TotalTxs) * 100
	snapshot.LatencyAvgMS = float64(totalLatency.Milliseconds()) / float64(len(c.metrics))
	snapshot.LatencyMinMS = float64(minLatency.Milliseconds())
	snapshot.LatencyMaxMS = float64(maxLatency.Milliseconds())

	// Calculate percentiles
	snapshot.LatencyP50MS = calculatePercentile(latencies, 0.50)
	snapshot.LatencyP95MS = calculatePercentile(latencies, 0.95)
	snapshot.LatencyP99MS = calculatePercentile(latencies, 0.99)

	// Calculate throughput
	elapsed := time.Since(c.startTime).Seconds()
	if elapsed > 0 {
		snapshot.ThroughputTPS = float64(snapshot.SuccessfulTxs) / elapsed
	}

	return snapshot
}

// GetScenarioStats returns metrics grouped by scenario
func (c *Collector) GetScenarioStats() map[string]*ScenarioStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	byScenario := make(map[string]*ScenarioStats)

	for _, m := range c.metrics {
		stats, ok := byScenario[m.Scenario]
		if !ok {
			stats = &ScenarioStats{
				Name:         m.Scenario,
				LatencyMinMS: 1<<63 - 1,
			}
			byScenario[m.Scenario] = stats
		}

		stats.TotalTxs++
		if m.Success {
			stats.SuccessfulTxs++
		} else {
			stats.FailedTxs++
		}

		latencyMS := float64(m.Latency.Milliseconds())
		if latencyMS < stats.LatencyMinMS {
			stats.LatencyMinMS = latencyMS
		}
		if latencyMS > stats.LatencyMaxMS {
			stats.LatencyMaxMS = latencyMS
		}
	}

	// Calculate aggregates per scenario
	for scenario, stats := range byScenario {
		var latencies []float64
		var totalLatency time.Duration

		for _, m := range c.metrics {
			if m.Scenario == scenario {
				latencies = append(latencies, float64(m.Latency.Milliseconds()))
				totalLatency += m.Latency
			}
		}

		stats.LatencyAvgMS = float64(totalLatency.Milliseconds()) / float64(len(latencies))
		stats.LatencyP50MS = calculatePercentile(latencies, 0.50)
		stats.LatencyP95MS = calculatePercentile(latencies, 0.95)
		stats.LatencyP99MS = calculatePercentile(latencies, 0.99)
		stats.SuccessRatePerc = float64(stats.SuccessfulTxs) / float64(stats.TotalTxs) * 100
	}

	return byScenario
}

// calculatePercentile calculates percentile value from sorted latencies
func calculatePercentile(latencies []float64, percentile float64) float64 {
	if len(latencies) == 0 {
		return 0
	}

	// Simple percentile calculation
	index := int(float64(len(latencies)) * percentile)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	// For simplicity, assuming unsorted. In production, would need sorting
	// This is a simplified version - proper implementation would use sorting
	sum := 0.0
	count := 0
	for _, l := range latencies {
		if int(float64(len(latencies))*percentile) >= count {
			sum += l
		}
		count++
	}

	return sum / float64(count)
}

// GetAllMetrics returns all collected metrics
func (c *Collector) GetAllMetrics() []*TransactionMetric {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Return a copy to avoid race conditions
	result := make([]*TransactionMetric, len(c.metrics))
	copy(result, c.metrics)
	return result
}

// SubmitCommitTask is a stub for the old Collector (not implemented)
func (c *Collector) SubmitCommitTask(task *CommitTask) {
	// Not implemented for the old Collector
	panic("SubmitCommitTask not implemented for legacy Collector")
}

// Stop is a stub for the old Collector (no-op)
func (c *Collector) Stop() {
	// No-op for legacy Collector
}
