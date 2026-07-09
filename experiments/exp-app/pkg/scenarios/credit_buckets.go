package scenarios

import (
	"context"
	"strings"
	"sync"

	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
)

// SharedCreditBuckets stores per-user credits partitioned between interop and
// bidding. One unreserved credit is added to interop queue per refresh; bidding
// uses the remaining available credits.
type SharedCreditBuckets struct {
	mu           sync.RWMutex
	interopQueue []*credits.MintCredit
	reserved     map[string]struct{}
	bidding      []*credits.MintCredit
	interopReady chan struct{}
}

// NewSharedCreditBuckets creates an empty shared bucket set.
func NewSharedCreditBuckets() *SharedCreditBuckets {
	return &SharedCreditBuckets{
		reserved:     make(map[string]struct{}),
		interopReady: make(chan struct{}, 1),
	}
}

// RefreshFromAvailable updates interop and bidding buckets from a fresh
// GetAvailableCreditsByOwner response.
func (s *SharedCreditBuckets) RefreshFromAvailable(available []credits.MintCredit) {
	s.mu.Lock()

	current := make(map[string]*credits.MintCredit, len(available))
	for _, credit := range available {
		if credit.Quantity <= 0 {
			continue
		}
		creditCopy := credit
		key, ok := mintCreditIDKey(&creditCopy)
		if !ok {
			continue
		}
		current[key] = &creditCopy
	}

	if len(current) == 0 {
		s.interopQueue = nil
		s.bidding = nil
		s.reserved = make(map[string]struct{})
		s.mu.Unlock()
		return
	}

	keptQueue := make([]*credits.MintCredit, 0, len(s.interopQueue))
	newReserved := make(map[string]struct{}, len(s.interopQueue)+1)
	for _, queued := range s.interopQueue {
		key, ok := mintCreditIDKey(queued)
		if !ok {
			continue
		}
		fresh, exists := current[key]
		if !exists {
			continue
		}
		keptQueue = append(keptQueue, fresh)
		newReserved[key] = struct{}{}
	}
	s.interopQueue = keptQueue

	for _, credit := range current {
		key, ok := mintCreditIDKey(credit)
		if !ok {
			continue
		}
		if _, isReserved := newReserved[key]; isReserved {
			continue
		}
		s.interopQueue = append(s.interopQueue, credit)
		newReserved[key] = struct{}{}
		break
	}

	bidding := make([]*credits.MintCredit, 0, len(current))
	for key, credit := range current {
		if _, isReserved := newReserved[key]; isReserved {
			continue
		}
		bidding = append(bidding, credit)
	}

	s.reserved = newReserved
	s.bidding = bidding
	hasInteropCredit := len(s.interopQueue) > 0
	s.mu.Unlock()

	if hasInteropCredit {
		s.signalInteropReady()
	}
}

// BiddingCredits returns current bidding credits.
func (s *SharedCreditBuckets) BiddingCredits() []*credits.MintCredit {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.bidding) == 0 {
		return nil
	}

	out := make([]*credits.MintCredit, len(s.bidding))
	copy(out, s.bidding)
	return out
}

// TakeInteropCredit pops next interop credit from queue.
func (s *SharedCreditBuckets) TakeInteropCredit(ctx context.Context) (*credits.MintCredit, bool) {
	for {
		s.mu.Lock()
		if len(s.interopQueue) > 0 {
			credit := s.interopQueue[0]
			s.interopQueue = s.interopQueue[1:]

			key, ok := mintCreditIDKey(credit)
			if ok {
				delete(s.reserved, key)
			}
			s.mu.Unlock()
			return credit, true
		}
		s.mu.Unlock()

		if err := s.drainInteropReady(ctx); err != nil {
			return nil, false
		}
	}
}

// signalInteropReady performs a non-blocking send to the interopReady channel
func (s *SharedCreditBuckets) signalInteropReady() {
	select {
	case s.interopReady <- struct{}{}:
	default:
	}
}

// drainInteropReady waits for a signal that interop credits are available or for the context to be done.
func (s *SharedCreditBuckets) drainInteropReady(ctx context.Context) error {
	select {
	case <-s.interopReady:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func mintCreditIDKey(credit *credits.MintCredit) (string, bool) {
	if credit == nil {
		return "", false
	}
	idParts := credit.GetID()
	if idParts == nil || len(*idParts) == 0 {
		panic("MintCredit has no ID parts")
	}
	return strings.Join((*idParts)[0], "|"), true
}
