package metrics

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// BucketSize defines the time window for each metrics bucket (100ms)
const (
	BucketSize       = 100 * time.Millisecond
	MaxBucketSamples = 1000 // Cap samples per bucket to avoid memory blowup
	DefaultWorkers   = 10   // Number of worker goroutines for commit awaiting
	CommitQueueSize  = 1000 // Buffered channel size for commit tasks
)

// LatencyBucket holds latency samples for a time bucket
type LatencyBucket struct {
	latencies []float64
	mu        sync.Mutex
}

type TransactionMetricBucket struct {
	metrics []*TransactionMetric
	mu      sync.Mutex
}

// ScenarioBucket tracks metrics for a scenario within a time bucket
type ScenarioBucket struct {
	name         string
	successCount int64
	failCount    int64
	latencies    *TransactionMetricBucket
}

// BucketedCollector uses time-bucketed metrics to avoid hot-path locking
type BucketedCollector struct {
	// Per-scenario, per-bucket metrics
	buckets     map[string]map[int64]*ScenarioBucket // scenario -> bucket_index -> bucket
	bucketsLock sync.RWMutex

	startTime           time.Time
	currentBucket       int64 // atomic: current bucket index
	bucketDuration      time.Duration
	timeWaitingForMutex int64 // atomic: microseconds spent waiting for locks

	// Worker pool for commit awaiting
	commitTasks chan *CommitTask
	workers     int
	wg          sync.WaitGroup
	done        chan struct{}
}

// NewBucketedCollector creates a new bucketed metrics collector
func NewBucketedCollector() *BucketedCollector {
	return NewBucketedCollectorWithWorkers(DefaultWorkers)
}

// NewBucketedCollectorWithWorkers creates a collector with a specified number of workers
func NewBucketedCollectorWithWorkers(workers int) *BucketedCollector {
	c := &BucketedCollector{
		buckets:        make(map[string]map[int64]*ScenarioBucket),
		startTime:      time.Now(),
		currentBucket:  0,
		bucketDuration: BucketSize,
		commitTasks:    make(chan *CommitTask, CommitQueueSize),
		workers:        workers,
		done:           make(chan struct{}),
	}
	c.startWorkers()
	return c
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

	// Measure lock contention for bucketsLock
	startLock := time.Now()
	c.bucketsLock.Lock()
	atomic.AddInt64(&c.timeWaitingForMutex, time.Since(startLock).Microseconds())

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
			latencies: &TransactionMetricBucket{metrics: make([]*TransactionMetric, 0, MaxBucketSamples)},
		}
		scenarioBuckets[bucketIdx] = bucket
	}
	c.bucketsLock.Unlock()

	// Record metric in bucket (short lock, sharded by scenario+bucket)
	startBucketLock := time.Now()
	bucket.latencies.mu.Lock()
	atomic.AddInt64(&c.timeWaitingForMutex, time.Since(startBucketLock).Microseconds())

	// Only append if under the safety cap (keep accurate throughput via atomic counters)
	if len(bucket.latencies.metrics) < MaxBucketSamples {
		// Store a shallow copy to ensure safety against caller mutation
		metricCopy := *m
		bucket.latencies.metrics = append(bucket.latencies.metrics, &metricCopy)
	}
	bucket.latencies.mu.Unlock()

	// Atomic increments (no lock needed, 100% accurate throughput)
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
	waitMicros := atomic.LoadInt64(&c.timeWaitingForMutex)
	fmt.Printf("Collector Lock Contention: %d us\n", waitMicros)

	for _, scenarioBuckets := range c.buckets {
		for _, bucket := range scenarioBuckets {
			successCount := atomic.LoadInt64(&bucket.successCount)
			failCount := atomic.LoadInt64(&bucket.failCount)

			snapshot.SuccessfulTxs += successCount
			snapshot.FailedTxs += failCount
			snapshot.TotalTxs += successCount + failCount

			// Collect latencies from this bucket
			bucket.latencies.mu.Lock()
			for _, m := range bucket.latencies.metrics {
				lat := float64(m.Latency.Milliseconds())
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
			// for _, lat := range bucket.latencies.latencies {
			for _, m := range bucket.latencies.metrics {
				lat := float64(m.Latency.Milliseconds())
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
	c.bucketsLock.RLock()
	defer c.bucketsLock.RUnlock()

	result := make([]*TransactionMetric, 0)
	for _, scenarioBuckets := range c.buckets {
		for _, bucket := range scenarioBuckets {
			bucket.latencies.mu.Lock()
			result = append(result, bucket.latencies.metrics...)
			bucket.latencies.mu.Unlock()
		}
	}
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

// startWorkers starts the worker goroutines for commit awaiting
func (c *BucketedCollector) startWorkers() {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.commitWorker()
	}
}

// commitWorker processes commit tasks from the channel
func (c *BucketedCollector) commitWorker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			return
		case task := <-c.commitTasks:
			c.processCommitTask(task)
		}
	}
}

// processCommitTask awaits a commit and records the metric
func (c *BucketedCollector) processCommitTask(task *CommitTask) {
	var finalErr error

	if task.TxErr != nil {
		finalErr = task.TxErr
	} else if task.Commit == nil {
		finalErr = fmt.Errorf("commit is nil")
	} else {
		status, commitErr := task.Commit.Status()
		if commitErr != nil {
			finalErr = fmt.Errorf("failed to get commit status: %w", commitErr)
		} else if status.Code != peer.TxValidationCode_VALID {
			finalErr = fmt.Errorf("failed due to %s", status.Code.String())
		} else if !status.Successful {
			finalErr = fmt.Errorf("transaction %s failed with code %s",
				status.TransactionID, status.Code)
		}

		if finalErr != nil && task.RunOnError != nil {
			task.RunOnError()
		}
	}

	// Record the metric
	c.Record(&TransactionMetric{
		ID:       task.ID,
		Scenario: task.Scenario,
		Latency:  time.Since(task.Start),
		Success:  finalErr == nil,
		Error:    getErrorString(finalErr),
	})
}

// SubmitCommitTask submits a commit task to the worker pool
func (c *BucketedCollector) SubmitCommitTask(task *CommitTask) {
	// Check if stopped first
	select {
	case <-c.done:
		// Collector is stopped, process synchronously
		c.processCommitTask(task)
		return
	default:
	}

	// Try to submit to workers
	select {
	case c.commitTasks <- task:
		// Task submitted successfully
	default:
		// Channel is full, handle asynchronously to avoid blocking
		go c.processCommitTask(task)
	}
}

// Stop gracefully shuts down the worker pool
func (c *BucketedCollector) Stop() {
	close(c.done)
	c.wg.Wait()
	close(c.commitTasks)
}

// getErrorString returns the error message or an empty string if nil
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
