package scenarios

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// InteropScenario manages cross-chain interoperability operations
type InteropScenario struct {
	executor  *workload.Executor
	collector metrics.MetricsCollector
}

// NewInteropScenario creates a new interop scenario
func NewInteropScenario(executor *workload.Executor) *InteropScenario {
	return &InteropScenario{
		executor:  executor,
		collector: executor.GetCollector(),
	}
}

// HTLCWorkflow executes a full HTLC cycle: LockCredit -> CreateHTLC -> ClaimHTLC
func (s *InteropScenario) HTLCWorkflow(ctx context.Context, carbonClient *gateway.ClientWrapper, interopClient *gateway.ClientWrapper, count int) error {
	ownerID := carbonClient.GetIdentityID()

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		creditsRes, err := carbonClient.EvaluateTransaction("GetAvailableCreditsByOwner", ownerID)
		if err != nil {
			continue
		}

		availableCredits := []credits.MintCredit{}
		if err := json.Unmarshal(creditsRes, &availableCredits); err != nil {
			continue
		}
		if len(availableCredits) == 0 {
			continue
		}

		credit := availableCredits[i%len(availableCredits)]
		creditID := (*credit.GetID())[0]
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
		recordTransaction(s.collector, fmt.Sprintf("interop-lock-%d", i), "interop-lock-credit", start, err)

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
		recordTransaction(s.collector, fmt.Sprintf("interop-htlc-create-%d", i), "interop-create-htlc", start, err)

		if err != nil {
			continue
		}

		// 3. ClaimHTLC on Interop Chain
		start = time.Now()
		_, err = interopClient.SubmitTransaction("ClaimHTLC", lockID, preimage)
		recordTransaction(s.collector, fmt.Sprintf("interop-htlc-claim-%d", i), "interop-claim-htlc", start, err)
	}

	return nil
}
