package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/gateway"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/metrics"
	"github.com/johannww/phd-impl/experiments/exp-app/pkg/workload"
)

// BiddingScenario creates buy and sell bids for auctions
type BiddingScenario struct {
	executor  *workload.Executor
	collector metrics.MetricsCollector
}

// NewBiddingScenario creates a new bidding scenario
func NewBiddingScenario(executor *workload.Executor) *BiddingScenario {
	return &BiddingScenario{
		executor:  executor,
		collector: executor.GetCollector(),
	}
}

// CreateSellBidsContinuous runs a continuous sell bidding loop based on available credits
func (s *BiddingScenario) CreateSellBidsContinuous(ctx context.Context, client *gateway.ClientWrapper, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	i := 0
	sellerID := client.GetIdentityID()

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
			}
			err = json.Unmarshal(creditsRes, &creditsList)
			if err != nil {
				log.Printf("Error unmarshaling credits for seller %s: %v", sellerID, err)
			}

			if len(creditsList) == 0 {
				continue
			}

			// 2. Create sell bid for available credits

			for _, credit := range creditsList {
				price := int64(20)

				transient := map[string][]byte{
					"price": []byte(strconv.FormatInt(price, 10)),
				}

				creditID := (*credit.GetID())[0]
				creditIDStr, err := json.Marshal(creditID)
				if err != nil {
					log.Fatalf("Error marshaling credit ID for credit %v: %v", creditID, err)
				}

				start := time.Now()
				_, txErr := client.SubmitWithTransient("CreateSellBidFromCredit", transient, strconv.FormatInt(credit.Quantity, 10), string(creditIDStr))
				recordTransaction(s.collector, fmt.Sprintf("sell-bid-cont-%d", i), "bidding-sell-continuous", start, txErr)
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
			quantity := int64(150)
			price := int64(25)

			transient := map[string][]byte{
				"price":    []byte(strconv.FormatInt(price, 10)),
				"quantity": []byte(strconv.FormatInt(quantity, 10)),
			}

			start := time.Now()
			_, txErr := client.SubmitWithTransient("CreateBuyBidPrivateQuantity", transient)
			recordTransaction(s.collector, fmt.Sprintf("buy-bid-cont-%d", i), "bidding-buy-continuous", start, txErr)

			i++
		}
	}
}
