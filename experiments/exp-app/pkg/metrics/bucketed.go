package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// BucketSize defines the time window for each metrics bucket (100ms)
const BucketSize = 100 * time.Millisecond

// LatencyBucket holds latency samples for a time bucket
type LatencyBucket struct {
	latencies []float64
	mu        sync.Mutex
}

// ScenarioBucket tracks metrics for a scenario within a time bucket
type ScenarioBucket struct {
	name         string
	successCount int64
	failCount    int64
	latencies    *LatencyBucket
}

// BucketedCollector uses time-bucketed metrics to avoid hot-path locking
type BucketedCollector struct {
	// Per-scenario, per-bucket metrics
	buckets     map[string]map[int64]*ScenarioBucket // scenario -> bucket_index -> bucket
	bucketsLock sync.RWMutex

	startTime      time.Time
	currentBucket  int64 // atomic: current bucket index
	bucketDuration time.Duration
}

// NewBucketedCollector creates a new bucketed metrics collector
func NewBucketedCollector() *BucketedCollector {
	return &BucketedCollector{
		buckets:        make(map[string]map[int64]*ScenarioBucket),
		startTime:      time.Now(),
		currentBucket:  0,
		bucketDuration: BucketSize,
	}
}

// getCurrentBucketIndex returns the bucket index for the current time
func (c *BucketedCollector) getCurrentBucketIndex() int64 {
	elapsed := time.Since(c.startTime)
	return int64(elapsed / c.bucketDuration)
}

// Record adds a transaction metric (minimal locking, sharded by scenario)
func (c *BucketedCollector) Record(m *TransactionMetric) {
	m.Timestamp = time.Now()

	bucketIdx := c.getCurrentBucketIndex()
	atomic.StoreInt64(&c.currentBucket, bucketIdx)

	// Get or create scenario's bucket map (rare operation, quick lock)
	c.bucketsLock.Lock()
	scenarioBuckets, ok := c.buckets[m.Scenario]
	if !ok {
		scenarioBuckets = make(map[int64]*ScenarioBucket)
		c.buckets[m.Scenario] = scenarioBuckets
	}

	// Get or create bucket for this scenario (quick, small lock)
	bucket, ok := scenarioBuckets[bucketIdx]
	if !ok {
		bucket = &ScenarioBucket{
			name:      m.Scenario,
			latencies: &LatencyBucket{latencies: make([]float64, 0, 256)},
		}
		scenarioBuckets[bucketIdx] = bucket
	}
	c.bucketsLock.Unlock()

	// Record metric in bucket (short lock, sharded by scenario+bucket)
	latencyMS := float64(m.Latency.Milliseconds())
	bucket.latencies.mu.Lock()
	bucket.latencies.latencies = append(bucket.latencies.latencies, latencyMS)
	bucket.latencies.mu.Unlock()

	// Atomic increments (no lock needed)
	if m.Success {
		atomic.AddInt64(&bucket.successCount, 1)
	} else {
		atomic.AddInt64(&bucket.failCount, 1)
	}
}

// GetSnapshot aggregates all buckets into a snapshot
func (c *BucketedCollector) GetSnapshot() *Snapshot {
	c.bucketsLock.RLock()
	defer c.bucketsLock.RUnlock()

	snapshot := &Snapshot{
		Timestamp: time.Now(),
	}

	allLatencies := make([]float64, 0, 10000)
	totalLatency := float64(0)

	for _, scenarioBuckets := range c.buckets {
		for _, bucket := range scenarioBuckets {
			successCount := atomic.LoadInt64(&bucket.successCount)
			failCount := atomic.LoadInt64(&bucket.failCount)

			snapshot.SuccessfulTxs += successCount
			snapshot.FailedTxs += failCount
			snapshot.TotalTxs += successCount + failCount

			// Collect latencies from this bucket
			bucket.latencies.mu.Lock()
			for _, lat := range bucket.latencies.latencies {
				allLatencies = append(allLatencies, lat)
				totalLatency += lat
			}
			bucket.latencies.mu.Unlock()
		}
	}

	if snapshot.TotalTxs == 0 {
		return snapshot
	}

	snapshot.SuccessRatePerc = float64(snapshot.SuccessfulTxs) / float64(snapshot.TotalTxs) * 100

	// Calculate latency statistics
	snapshot.LatencyAvgMS = totalLatency / float64(snapshot.TotalTxs)

	if len(allLatencies) > 0 {
		sort.Float64s(allLatencies)
		snapshot.LatencyMinMS = allLatencies[0]
		snapshot.LatencyMaxMS = allLatencies[len(allLatencies)-1]
		snapshot.LatencyP50MS = calculatePercentileFromSorted(allLatencies, 0.50)
		snapshot.LatencyP95MS = calculatePercentileFromSorted(allLatencies, 0.95)
		snapshot.LatencyP99MS = calculatePercentileFromSorted(allLatencies, 0.99)
	}

	// Calculate throughput
	elapsed := time.Since(c.startTime).Seconds()
	if elapsed > 0 {
		snapshot.ThroughputTPS = float64(snapshot.SuccessfulTxs) / elapsed
	}

	return snapshot
}

// GetScenarioStats returns metrics grouped by scenario
func (c *BucketedCollector) GetScenarioStats() map[string]*ScenarioStats {
	c.bucketsLock.RLock()
	defer c.bucketsLock.RUnlock()

	byScenario := make(map[string]*ScenarioStats)

	for scenario, scenarioBuckets := range c.buckets {
		stats := &ScenarioStats{
			Name:         scenario,
			LatencyMinMS: 1<<63 - 1,
		}

		allLatencies := make([]float64, 0)
		totalLatency := float64(0)

		for _, bucket := range scenarioBuckets {
			successCount := atomic.LoadInt64(&bucket.successCount)
			failCount := atomic.LoadInt64(&bucket.failCount)

			stats.TotalTxs += successCount + failCount
			stats.SuccessfulTxs += successCount
			stats.FailedTxs += failCount

			// Collect latencies
			bucket.latencies.mu.Lock()
			for _, lat := range bucket.latencies.latencies {
				allLatencies = append(allLatencies, lat)
				totalLatency += lat
				if lat < stats.LatencyMinMS {
					stats.LatencyMinMS = lat
				}
				if lat > stats.LatencyMaxMS {
					stats.LatencyMaxMS = lat
				}
			}
			bucket.latencies.mu.Unlock()
		}

		if stats.TotalTxs > 0 {
			stats.LatencyAvgMS = totalLatency / float64(stats.TotalTxs)
			stats.SuccessRatePerc = float64(stats.SuccessfulTxs) / float64(stats.TotalTxs) * 100

			if len(allLatencies) > 0 {
				sort.Float64s(allLatencies)
				stats.LatencyP50MS = calculatePercentileFromSorted(allLatencies, 0.50)
				stats.LatencyP95MS = calculatePercentileFromSorted(allLatencies, 0.95)
				stats.LatencyP99MS = calculatePercentileFromSorted(allLatencies, 0.99)
			}
		}

		byScenario[scenario] = stats
	}

	return byScenario
}

// GetAllMetrics returns all collected metrics (expensive, for export)
func (c *BucketedCollector) GetAllMetrics() []*TransactionMetric {
	// Note: bucketed design doesn't store individual TransactionMetric pointers
	// Return aggregated view as a reconstructed list (for compatibility)
	c.bucketsLock.RLock()
	defer c.bucketsLock.RUnlock()

	result := make([]*TransactionMetric, 0)
	// This is a compatibility shim; ideally callers would use GetSnapshot/GetScenarioStats
	// For now, return empty (individual metrics are not preserved)
	return result
}

// calculatePercentileFromSorted calculates percentile from a sorted float64 slice
func calculatePercentileFromSorted(sorted []float64, percentile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)-1) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	if index < 0 {
		index = 0
	}
	return sorted[index]
}
