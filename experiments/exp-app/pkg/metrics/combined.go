package metrics

import (
	"sort"
	"time"
)

// CombinedCollector merges multiple collectors into one reporting view.
type CombinedCollector struct {
	collectors []MetricsCollector
}

// NewCombinedCollector creates a collector that aggregates all provided collectors.
func NewCombinedCollector(collectors ...MetricsCollector) *CombinedCollector {
	filtered := make([]MetricsCollector, 0, len(collectors))
	for _, collector := range collectors {
		if collector != nil {
			filtered = append(filtered, collector)
		}
	}

	return &CombinedCollector{collectors: filtered}
}

// Record broadcasts a metric to all underlying collectors.
func (c *CombinedCollector) Record(m *TransactionMetric) {
	for _, collector := range c.collectors {
		collector.Record(m)
	}
}

// GetSnapshot returns an aggregated snapshot over all collectors.
func (c *CombinedCollector) GetSnapshot() *Snapshot {
	allMetrics := c.GetAllMetrics()
	return snapshotFromMetrics(allMetrics)
}

// GetScenarioStats returns merged scenario stats over all collectors.
func (c *CombinedCollector) GetScenarioStats() map[string]*ScenarioStats {
	allMetrics := c.GetAllMetrics()
	return scenarioStatsFromMetrics(allMetrics)
}

// GetAllMetrics returns all metrics from all collectors.
func (c *CombinedCollector) GetAllMetrics() []*TransactionMetric {
	all := make([]*TransactionMetric, 0)
	for _, collector := range c.collectors {
		all = append(all, collector.GetAllMetrics()...)
	}
	return all
}

// SubmitCommitTask broadcasts commit tasks to all underlying collectors.
func (c *CombinedCollector) SubmitCommitTask(task *CommitTask) {
	for _, collector := range c.collectors {
		collector.SubmitCommitTask(task)
	}
}

// Stop stops all underlying collectors.
func (c *CombinedCollector) Stop() {
	for _, collector := range c.collectors {
		collector.Stop()
	}
}

func snapshotFromMetrics(metrics []*TransactionMetric) *Snapshot {
	snapshot := &Snapshot{Timestamp: time.Now()}
	if len(metrics) == 0 {
		return snapshot
	}

	latencies := make([]float64, 0, len(metrics))
	var totalLatency float64
	var earliest time.Time
	var hasEarliest bool

	for _, metric := range metrics {
		snapshot.TotalTxs++
		if metric.Success {
			snapshot.SuccessfulTxs++
		} else {
			snapshot.FailedTxs++
		}

		lat := float64(metric.Latency.Milliseconds())
		latencies = append(latencies, lat)
		totalLatency += lat

		if !hasEarliest || metric.Timestamp.Before(earliest) {
			earliest = metric.Timestamp
			hasEarliest = true
		}
	}

	snapshot.SuccessRatePerc = float64(snapshot.SuccessfulTxs) / float64(snapshot.TotalTxs) * 100
	snapshot.LatencyAvgMS = totalLatency / float64(len(latencies))

	sort.Float64s(latencies)
	snapshot.LatencyMinMS = latencies[0]
	snapshot.LatencyMaxMS = latencies[len(latencies)-1]
	snapshot.LatencyP50MS = percentileFromSorted(latencies, 0.50)
	snapshot.LatencyP95MS = percentileFromSorted(latencies, 0.95)
	snapshot.LatencyP99MS = percentileFromSorted(latencies, 0.99)

	if hasEarliest {
		elapsed := time.Since(earliest).Seconds()
		if elapsed > 0 {
			snapshot.ThroughputTPS = float64(snapshot.SuccessfulTxs) / elapsed
		}
	}

	return snapshot
}

func scenarioStatsFromMetrics(metrics []*TransactionMetric) map[string]*ScenarioStats {
	byScenario := make(map[string][]*TransactionMetric)
	for _, metric := range metrics {
		byScenario[metric.Scenario] = append(byScenario[metric.Scenario], metric)
	}

	result := make(map[string]*ScenarioStats, len(byScenario))
	for scenario, scenarioMetrics := range byScenario {
		stats := &ScenarioStats{Name: scenario}
		latencies := make([]float64, 0, len(scenarioMetrics))
		var totalLatency float64

		for _, metric := range scenarioMetrics {
			stats.TotalTxs++
			if metric.Success {
				stats.SuccessfulTxs++
			} else {
				stats.FailedTxs++
			}

			lat := float64(metric.Latency.Milliseconds())
			latencies = append(latencies, lat)
			totalLatency += lat
		}

		if len(latencies) > 0 {
			sort.Float64s(latencies)
			stats.LatencyAvgMS = totalLatency / float64(len(latencies))
			stats.LatencyMinMS = latencies[0]
			stats.LatencyMaxMS = latencies[len(latencies)-1]
			stats.LatencyP50MS = percentileFromSorted(latencies, 0.50)
			stats.LatencyP95MS = percentileFromSorted(latencies, 0.95)
			stats.LatencyP99MS = percentileFromSorted(latencies, 0.99)
		}

		if stats.TotalTxs > 0 {
			stats.SuccessRatePerc = float64(stats.SuccessfulTxs) / float64(stats.TotalTxs) * 100
		}

		result[scenario] = stats
	}

	return result
}

func percentileFromSorted(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if q <= 0 {
		return sorted[0]
	}
	if q >= 1 {
		return sorted[len(sorted)-1]
	}

	idx := int(float64(len(sorted)-1) * q)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
