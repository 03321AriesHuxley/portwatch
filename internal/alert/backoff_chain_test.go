package alert

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackoffWithJitter verifies that combining JitterNotifier and
// BackoffNotifier produces correct retry behaviour with spread delays.
func TestBackoffWithJitter(t *testing.T) {
	var calls int32
	sentinel := errors.New("fail")
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			if atomic.AddInt32(&calls, 1) < 3 {
				return sentinel
			}
			return nil
		},
	}
	// Wrap inner with jitter, then wrap that with backoff.
	jittered := NewJitterNotifier(inner, 1*time.Millisecond, 5*time.Millisecond)
	n := NewBackoffNotifier(jittered, 5,
		WithInitialDelay(5*time.Millisecond),
		WithMaxDelay(20*time.Millisecond),
	)
	err := n.Send(context.Background(), makeBackoffEvents())
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

// TestJitterWithBackoff verifies the reverse composition: backoff inside jitter.
func TestJitterWithBackoff(t *testing.T) {
	var calls int32
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			atomic.AddInt32(&calls, 1)
			return nil
		},
	}
	backed := NewBackoffNotifier(inner, 3,
		WithInitialDelay(5*time.Millisecond),
	)
	n := NewJitterNotifier(backed, 1*time.Millisecond, 3*time.Millisecond)
	err := n.Send(context.Background(), makeJitterEvents())
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

// TestBackoffWithJitter_ContextCancelStopsChain ensures cancellation propagates
// through the full chain.
func TestBackoffWithJitter_ContextCancelStopsChain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			return errors.New("always fails")
		},
	}
	jittered := NewJitterNotifier(inner, 1*time.Millisecond, 3*time.Millisecond)
	n := NewBackoffNotifier(jittered, 100,
		WithInitialDelay(1*time.Millisecond),
		WithMaxDelay(5*time.Millisecond),
	)
	err := n.Send(ctx, makeBackoffEvents())
	require.Error(t, err)
}
