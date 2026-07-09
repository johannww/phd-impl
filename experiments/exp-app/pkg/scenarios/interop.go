package scenarios

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// InteropScenario manages cross-chain interoperability operations
type InteropScenario struct {
	executor  *workload.Executor
	collector metrics.MetricsCollector
	buckets   *SharedCreditBuckets
}

// NewInteropScenario creates a new interop scenario
func NewInteropScenario(executor *workload.Executor, buckets *SharedCreditBuckets) *InteropScenario {
	return &InteropScenario{
		executor:  executor,
		collector: executor.GetCollector(),
		buckets:   buckets,
	}
}

// HTLCWorkflow continuously executes HTLC cycles until the context is cancelled:
// LockCredit -> CreateHTLC -> ClaimHTLC.
func (s *InteropScenario) HTLCWorkflow(ctx context.Context, carbonClient *gateway.ClientWrapper, interopClient *gateway.ClientWrapper) error {
	if s.buckets == nil {
		return fmt.Errorf("shared credit buckets are required for interop workflow")
	}

	ownerID := carbonClient.GetIdentityID()

	for attempt := 0; ; attempt++ {
		assignedCredit, ok := s.buckets.TakeInteropCredit(ctx)
		if !ok {
			if err := ctx.Err(); err != nil {
				return err
			}
			continue
		}

		idParts := assignedCredit.GetID()
		if idParts == nil || len(*idParts) == 0 {
			panic(fmt.Sprintf("credit has no ID parts: %v", assignedCredit))
		}
		creditID := (*idParts)[0]

		creditIDBytes, err := json.Marshal(creditID)
		if err != nil {
			continue
		}

		destChainID := "target-chain"
		preimage := fmt.Sprintf("interop-preimage-%d", time.Now().UnixNano())
		hashLockBytes := sha256.Sum256([]byte(preimage))
		hashLock := hex.EncodeToString(hashLockBytes[:])
		validUntil := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
		amount := "100"

		// 1. LockCredit on Carbon Chain
		start := time.Now()
		res, err := carbonClient.SubmitTransaction(
			"LockCredit",
			string(creditIDBytes),
			"0",
			destChainID,
		)
		lockID := string(res)
		recordTransaction(s.collector, fmt.Sprintf("interop-lock-%d", attempt), "interop-lock-credit", start, err)

		if err != nil {
			continue
		}

		// 2. CreateHTLC on Interop Chain
		start = time.Now()
		_, err = interopClient.SubmitTransaction(
			"CreateHTLC",
			string(creditIDBytes),
			lockID,
			hashLock,
			ownerID,
			validUntil,
			destChainID,
			amount,
		)
		recordTransaction(s.collector, fmt.Sprintf("interop-htlc-create-%d", attempt), "interop-create-htlc", start, err)

		if err != nil {
			continue
		}

		// 3. ClaimHTLC on Interop Chain
		start = time.Now()
		_, err = interopClient.SubmitTransaction("ClaimHTLC", lockID, preimage)
		recordTransaction(s.collector, fmt.Sprintf("interop-htlc-claim-%d", attempt), "interop-claim-htlc", start, err)
	}

	return nil
}
