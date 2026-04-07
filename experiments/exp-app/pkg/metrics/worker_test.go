package metrics

import (
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	collector := NewBucketedCollector()
	defer collector.Stop()

	// Submit a task with a transaction error
	task := &CommitTask{
		ID:       "test-1",
		Scenario: "test",
		Start:    time.Now(),
		Commit:   nil,
		TxErr:    nil, // No error, but no commit either (will be handled)
	}

	collector.SubmitCommitTask(task)

	// Give workers time to process
	time.Sleep(100 * time.Millisecond)

	// Verify the metric was recorded (even with nil commit)
	snapshot := collector.GetSnapshot()
	if snapshot.TotalTxs < 1 {
		t.Errorf("expected at least 1 transaction, got %d", snapshot.TotalTxs)
	}
}

func TestWorkerPoolConcurrency(t *testing.T) {
	collector := NewBucketedCollectorWithWorkers(5)
	defer collector.Stop()

	// Submit many tasks concurrently
	numTasks := 100
	for i := 0; i < numTasks; i++ {
		task := &CommitTask{
			ID:       "test",
			Scenario: "concurrent",
			Start:    time.Now(),
			Commit:   nil,
			TxErr:    nil,
		}
		collector.SubmitCommitTask(task)
	}

	// Give workers time to process all tasks
	time.Sleep(500 * time.Millisecond)

	snapshot := collector.GetSnapshot()
	if snapshot.TotalTxs < int64(numTasks) {
		t.Errorf("expected at least %d transactions, got %d", numTasks, snapshot.TotalTxs)
	}
}

func TestWorkerPoolStop(t *testing.T) {
	collector := NewBucketedCollector()

	// Submit a task
	task := &CommitTask{
		ID:       "test",
		Scenario: "stop-test",
		Start:    time.Now(),
		Commit:   nil,
		TxErr:    nil,
	}
	collector.SubmitCommitTask(task)

	// Stop should wait for workers to finish
	collector.Stop()

	// After stop, submitting should not panic (handled by fallback)
	collector.SubmitCommitTask(task)

	snapshot := collector.GetSnapshot()
	if snapshot.TotalTxs < 1 {
		t.Errorf("expected at least 1 transaction before stop, got %d", snapshot.TotalTxs)
	}
}
