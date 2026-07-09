package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/johannww/phd-impl/chaincodes/carbon/credits"
	"github.com/stretchr/testify/require"
)

func TestSharedCreditBuckets_TakeInteropCreditWaitsUntilAvailable(t *testing.T) {
	buckets := NewSharedCreditBuckets()

	creditCh := make(chan *credits.MintCredit, 1)
	okCh := make(chan bool, 1)
	go func() {
		credit, ok := buckets.TakeInteropCredit(context.Background())
		creditCh <- credit
		okCh <- ok
	}()

	select {
	case credit := <-creditCh:
		t.Fatalf("take returned too early with credit: %#v", credit)
	case <-time.After(50 * time.Millisecond):
	}

	buckets.RefreshFromAvailable([]credits.MintCredit{{
		Credit: credits.Credit{
			OwnerID:  "owner",
			ChunkID:  []string{"chunk-1"},
			Quantity: 100,
		},
		MintTimeStamp: "credit-1",
	}})

	select {
	case credit := <-creditCh:
		require.NotNil(t, credit)
		require.True(t, <-okCh)
	case <-time.After(time.Second):
		t.Fatal("take did not unblock after credits were refreshed")
	}

	require.Zero(t, len(buckets.interopQueue))
	require.Zero(t, len(buckets.interopReady))
}

func TestSharedCreditBuckets_TakeInteropCreditReturnsFalseOnContextCancel(t *testing.T) {
	buckets := NewSharedCreditBuckets()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	credit, ok := buckets.TakeInteropCredit(ctx)
	require.Nil(t, credit)
	require.False(t, ok)
}
