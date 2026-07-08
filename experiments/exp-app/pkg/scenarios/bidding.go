package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// BiddingScenario creates buy and sell bids for auctions
type BiddingScenario struct {
	executor        *workload.Executor
	collector       metrics.MetricsCollector
	buckets         *SharedCreditBuckets
	walletsPerBuyer int
}

// NewBiddingScenario creates a new bidding scenario
func NewBiddingScenario(executor *workload.Executor, buckets *SharedCreditBuckets, walletsPerBuyer int) *BiddingScenario {
	if walletsPerBuyer <= 0 {
		walletsPerBuyer = 1
	}

	return &BiddingScenario{
		executor:        executor,
		collector:       executor.GetCollector(),
		buckets:         buckets,
		walletsPerBuyer: walletsPerBuyer,
	}
}

// CreateSellBidsContinuous runs a continuous sell bidding loop based on available credits
func (s *BiddingScenario) CreateSellBidsContinuous(ctx context.Context, client *gateway.ClientWrapper, interval time.Duration) error {
	if s.buckets == nil {
		return fmt.Errorf("shared credit buckets are required for bidding scenario")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	i := 0
	sellerID := client.GetIdentityID()
	usedCredits := &sync.Map{}
	runOnErrorGenerator := func(creditIDStr string) RunOnErrorFunc {
		return func() {
			usedCredits.Store(creditIDStr, false)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// 1. Fetch available credits
			creditsList := []credits.MintCredit{}
			creditsRes, err := client.EvaluateTransaction("GetAvailableCreditsByOwner", sellerID)
			if err != nil {
				log.Printf("Error fetching credits for seller %s: %v", sellerID, err)
				continue
			}
			err = json.Unmarshal(creditsRes, &creditsList)
			if err != nil {
				log.Printf("Error unmarshaling credits for seller %s: %v", sellerID, err)
				continue
			}
			s.buckets.RefreshFromAvailable(creditsList)

			if len(creditsList) == 0 {
				continue
			}

			// 2. Create sell bid for available credits
			biddingCredits := s.buckets.BiddingCredits()

			for _, credit := range biddingCredits {
				select {
				case <-ctx.Done():
					log.Printf("Context done while creating sell bids: %v", ctx.Err())
					return ctx.Err()
				default:
				}

				idParts := credit.GetID()
				if idParts == nil || len(*idParts) == 0 {
					panic(fmt.Sprintf("Credit %v has no ID parts", credit))
				}
				creditID := (*idParts)[0]
				creditIDBytes, err := json.Marshal(creditID)
				if err != nil {
					log.Fatalf("Error marshaling credit ID for credit %v: %v", creditID, err)
				}
				creditIDStr := string(creditIDBytes)

				if val, ok := usedCredits.Load(creditIDStr); ok && val.(bool) {
					// fmt.Printf("Credit %s already. %t\n", creditIDStr, val.(bool))
					continue
				}
				// fmt.Printf("Using credit %s to create sell bid.\n", creditIDStr)
				usedCredits.Store(creditIDStr, true)

				price := int64(20)

				transient := map[string][]byte{
					"price": []byte(strconv.FormatInt(price, 10)),
				}

				start := time.Now()
				_, commit, txErr := client.SubmitAsyncWithTransient("CreateSellBidFromCredit", transient, strconv.FormatInt(credit.Quantity, 10), creditIDStr)
				awaitAndRecord(s.collector, fmt.Sprintf("sell-bid-cont-%d", i), "bidding-sell-continuous", start, commit, txErr, runOnErrorGenerator(creditIDStr))
				i++

			}

		}
	}
}

// CreateBuyBidsContinuous runs a continuous buy bidding loop
func (s *BiddingScenario) CreateBuyBidsContinuous(ctx context.Context, client *gateway.ClientWrapper, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	i := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			quantity := int64(2000) // must be greater than common.QUANTITY_SCALE
			price := int64(25)
			walletNumber := i % s.walletsPerBuyer

			transient := map[string][]byte{
				"price":    []byte(strconv.FormatInt(price, 10)),
				"quantity": []byte(strconv.FormatInt(quantity, 10)),
			}

			start := time.Now()
			_, commit, txErr := client.SubmitAsyncWithTransient("CreateBuyBidPrivateQuantityForWallet", transient, strconv.Itoa(walletNumber))
			awaitAndRecord(s.collector, fmt.Sprintf("buy-bid-cont-%d", i), "bidding-buy-continuous", start, commit, txErr, nil)

			i++
		}
	}
}
