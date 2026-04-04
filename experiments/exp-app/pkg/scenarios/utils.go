package scenarios

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
)

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

// awaitAndRecord waits for a Fabric async commit in a new goroutine and records
// the end-to-end latency (from start) once the block is committed.
// It is safe to call from any scenario: the goroutine owns all its arguments.
func awaitAndRecord(collector metrics.MetricsCollector, id string, scenario string, start time.Time, commit *client.Commit, txErr error) {
	go func() {
		if txErr != nil {
			recordTransaction(collector, id, scenario, start, txErr)
			return
		}
		status, commitErr := commit.Status()
		if commitErr != nil {
			commitErr = fmt.Errorf("failed to get commit status: %w", commitErr)
		} else if status.Code != peer.TxValidationCode_VALID {
			commitErr = fmt.Errorf("failed due to %s", status.Code.String())
		} else if !status.Successful {
			commitErr = fmt.Errorf("transaction %s failed with code %s",
				status.TransactionID, status.Code)
		}
		recordTransaction(collector, id, scenario, start, commitErr)
	}()
}
