package scenarios

import (
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
)

type RunOnErrorFunc func()

// getErrorString returns the error message or an empty string if nil.
// This is shared across all scenario files in this package.
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// recordTransaction creates and records a TransactionMetric for a scenario.
func recordTransaction(collector metrics.MetricsCollector, id string, scenario string, start time.Time, err error) {
	collector.Record(&metrics.TransactionMetric{
		ID:       id,
		Scenario: scenario,
		Latency:  time.Since(start),
		Success:  err == nil,
		Error:    getErrorString(err),
	})
}

// awaitAndRecord submits a commit awaiting task to the metrics collector's worker pool.
// This avoids creating a goroutine per transaction, using a fixed worker pool instead.
func awaitAndRecord(
	collector metrics.MetricsCollector,
	id string,
	scenario string,
	start time.Time,
	commit *client.Commit,
	txErr error,
	runOnError RunOnErrorFunc) {

	collector.SubmitCommitTask(&metrics.CommitTask{
		ID:         id,
		Scenario:   scenario,
		Start:      start,
		Commit:     commit,
		TxErr:      txErr,
		RunOnError: runOnError,
	})
}
