package fixtures

import (
	"fmt"
	"math/rand"
	"time"
)

// BidData represents test data for a bid
type BidData struct {
	Quantity int64
	Price    int64
	UserID   string
}

// CreditData represents test data for a credit
type CreditData struct {
	Quantity   int64
	PropertyID string
	OwnerID    string
}

// AuctionData represents test data for an auction
type AuctionData struct {
	AuctionID string
	BuyBids   []BidData
	SellBids  []BidData
}

// Generator creates test fixtures
type Generator struct {
	seed int64
	rng  *rand.Rand
}

// NewGenerator creates a new fixture generator
func NewGenerator(seed int64) *Generator {
	rng := rand.New(rand.NewSource(seed))
	return &Generator{
		seed: seed,
		rng:  rng,
	}
}

// GenerateBuyBids generates N buy bids with random quantities
func (g *Generator) GenerateBuyBids(count int) []BidData {
	bids := make([]BidData, count)
	for i := 0; i < count; i++ {
		bids[i] = BidData{
			Quantity: int64(g.rng.Intn(1000) + 100), // 100-1100
			Price:    int64(g.rng.Intn(500) + 50),   // 50-550
			UserID:   fmt.Sprintf("buyer-%d", i),
		}
	}
	return bids
}

// GenerateSellBids generates N sell bids with random quantities
func (g *Generator) GenerateSellBids(count int) []BidData {
	bids := make([]BidData, count)
	for i := 0; i < count; i++ {
		bids[i] = BidData{
			Quantity: int64(g.rng.Intn(800) + 50), // 50-850
			Price:    int64(g.rng.Intn(400) + 20), // 20-420
			UserID:   fmt.Sprintf("seller-%d", i),
		}
	}
	return bids
}

// GenerateCredits generates N credits with random quantities
func (g *Generator) GenerateCredits(count int) []CreditData {
	credits := make([]CreditData, count)
	for i := 0; i < count; i++ {
		credits[i] = CreditData{
			Quantity:   int64(g.rng.Intn(5000) + 1000), // 1000-6000
			PropertyID: fmt.Sprintf("property-%d", i),
			OwnerID:    fmt.Sprintf("owner-%d", i%10), // Reuse some owners
		}
	}
	return credits
}

// GenerateAuction generates a complete auction with matched bids
func (g *Generator) GenerateAuction(auctionID string, buyBidCount, sellBidCount int) *AuctionData {
	return &AuctionData{
		AuctionID: auctionID,
		BuyBids:   g.GenerateBuyBids(buyBidCount),
		SellBids:  g.GenerateSellBids(sellBidCount),
	}
}

// RandomDelay returns a random delay for realistic load patterns
// Uses exponential distribution for inter-arrival times
func (g *Generator) RandomDelay(meanMillis int) time.Duration {
	// Exponential distribution: rate = 1/mean
	rate := 1.0 / float64(meanMillis)
	delay := -1.0 / rate * float64(g.rng.Int63())
	if delay < 0 {
		delay = -delay
	}
	return time.Duration(int64(delay)) * time.Millisecond
}

// RandomScenario selects a random scenario based on weights
// Example: weights = map[string]float64{"bidding": 0.4, "credit": 0.3, "tee": 0.3}
func (g *Generator) RandomScenario(weights map[string]float64) string {
	r := g.rng.Float64()
	cumsum := 0.0

	for scenario, weight := range weights {
		cumsum += weight
		if r <= cumsum {
			return scenario
		}
	}

	// Fallback to first scenario
	for scenario := range weights {
		return scenario
	}
	return ""
}

// ShuffleUint64 shuffles a slice of uint64 values
func (g *Generator) Shuffle(slice []int64) {
	for i := len(slice) - 1; i > 0; i-- {
		j := g.rng.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
