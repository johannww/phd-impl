package metrics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// Reporter handles output of metrics in various formats
type Reporter struct {
	collector MetricsCollector
}

// NewReporter creates a new reporter
func NewReporter(c MetricsCollector) *Reporter {
	return &Reporter{collector: c}
}

// PrintLiveMetrics prints current metrics to console
func (r *Reporter) PrintLiveMetrics() {
	snapshot := r.collector.GetSnapshot()
	fmt.Printf("\rTxs: %d/%d | Success: %.1f%% | TPS: %.2f | P50: %.0fms | P95: %.0fms | P99: %.0fms",
		snapshot.TotalTxs,
		snapshot.SuccessfulTxs,
		snapshot.SuccessRatePerc,
		snapshot.ThroughputTPS,
		snapshot.LatencyP50MS,
		snapshot.LatencyP95MS,
		snapshot.LatencyP99MS,
	)
}

// PrintFinalReport prints final metrics report
func (r *Reporter) PrintFinalReport() {
	snapshot := r.collector.GetSnapshot()
	stats := r.collector.GetScenarioStats()

	fmt.Println("\n=== FINAL PERFORMANCE REPORT ===")

	fmt.Printf("Total Transactions: %d\n", snapshot.TotalTxs)
	fmt.Printf("Successful: %d (%.2f%%)\n", snapshot.SuccessfulTxs, snapshot.SuccessRatePerc)
	fmt.Printf("Failed: %d\n", snapshot.FailedTxs)
	fmt.Printf("Throughput: %.2f TPS\n\n", snapshot.ThroughputTPS)

	fmt.Println("Latency Statistics (ms):")
	fmt.Printf("  Average: %.2f\n", snapshot.LatencyAvgMS)
	fmt.Printf("  Min: %.2f\n", snapshot.LatencyMinMS)
	fmt.Printf("  Max: %.2f\n", snapshot.LatencyMaxMS)
	fmt.Printf("  P50: %.2f\n", snapshot.LatencyP50MS)
	fmt.Printf("  P95: %.2f\n", snapshot.LatencyP95MS)
	fmt.Printf("  P99: %.2f\n\n", snapshot.LatencyP99MS)

	fmt.Println("Per-Scenario Statistics:")
	for scenario, s := range stats {
		fmt.Printf("\n%s:\n", scenario)
		fmt.Printf("  Transactions: %d (Success: %.2f%%)\n", s.TotalTxs, s.SuccessRatePerc)
		fmt.Printf("  Latency P50/P95/P99: %.2f/%.2f/%.2f ms\n",
			s.LatencyP50MS, s.LatencyP95MS, s.LatencyP99MS)
	}
}

// ExportJSON exports metrics to JSON file
func (r *Reporter) ExportJSON(filename string) error {
	snapshot := r.collector.GetSnapshot()
	stats := r.collector.GetScenarioStats()

	data := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_transactions":   snapshot.TotalTxs,
			"successful":           snapshot.SuccessfulTxs,
			"failed":               snapshot.FailedTxs,
			"success_rate_percent": snapshot.SuccessRatePerc,
			"throughput_tps":       snapshot.ThroughputTPS,
		},
		"latency_ms": map[string]interface{}{
			"average": snapshot.LatencyAvgMS,
			"min":     snapshot.LatencyMinMS,
			"max":     snapshot.LatencyMaxMS,
			"p50":     snapshot.LatencyP50MS,
			"p95":     snapshot.LatencyP95MS,
			"p99":     snapshot.LatencyP99MS,
		},
		"by_scenario": stats,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	return os.WriteFile(filename, jsonBytes, 0o644)
}

// ExportCSV exports transaction metrics to CSV file
func (r *Reporter) ExportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "scenario", "timestamp", "latency_ms", "success", "error"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	// Write metrics
	allMetrics := r.collector.GetAllMetrics()
	slices.SortFunc(allMetrics, func(a, b *TransactionMetric) int {
		return a.Timestamp.Compare(b.Timestamp)
	})

	for _, m := range allMetrics {
		record := []string{
			m.ID,
			m.Scenario,
			m.Timestamp.Format("2006-01-02 15:04:05.000"),
			fmt.Sprintf("%.2f", float64(m.Latency.Milliseconds())),
			fmt.Sprintf("%v", m.Success),
			m.Error,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write csv record: %w", err)
		}
	}

	return nil
}
