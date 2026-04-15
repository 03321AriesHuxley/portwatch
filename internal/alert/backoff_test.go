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

func makeBackoffEvents() []Event {
	return []Event{{Kind: EventAdded, Port: 8080, Proto: "tcp", Addr: "0.0.0.0"}}
}

func TestBackoffNotifier_SucceedsFirstAttempt(t *testing.T) {
	var calls int32
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			atomic.AddInt32(&calls, 1)
			return nil
		},
	}
	n := NewBackoffNotifier(inner, 3,
		WithInitialDelay(10*time.Millisecond),
	)
	err := n.Send(context.Background(), makeBackoffEvents())
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestBackoffNotifier_RetriesAndSucceeds(t *testing.T) {
	var calls int32
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			n := atomic.AddInt32(&calls, 1)
			if n < 3 {
				return errors.New("transient")
			}
			return nil
		},
	}
	n := NewBackoffNotifier(inner, 5,
		WithInitialDelay(5*time.Millisecond),
		WithMaxDelay(50*time.Millisecond),
		WithMultiplier(2.0),
	)
	err := n.Send(context.Background(), makeBackoffEvents())
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestBackoffNotifier_ExhaustsAttempts(t *testing.T) {
	var calls int32
	sentinel := errors.New("always fails")
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			atomic.AddInt32(&calls, 1)
			return sentinel
		},
	}
	n := NewBackoffNotifier(inner, 3,
		WithInitialDelay(5*time.Millisecond),
	)
	err := n.Send(context.Background(), makeBackoffEvents())
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestBackoffNotifier_RespectsContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			if atomic.AddInt32(&calls, 1) == 1 {
				cancel()
			}
			return errors.New("fail")
		},
	}
	n := NewBackoffNotifier(inner, 10,
		WithInitialDelay(50*time.Millisecond),
	)
	err := n.Send(ctx, makeBackoffEvents())
	require.Error(t, err)
	// Should stop early due to context cancellation.
	assert.Less(t, atomic.LoadInt32(&calls), int32(10))
}

func TestBackoffNotifier_MaxDelayIsRespected(t *testing.T) {
	var calls int32
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			atomic.AddInt32(&calls, 1)
			return errors.New("fail")
		},
	}
	start := time.Now()
	n := NewBackoffNotifier(inner, 3,
		WithInitialDelay(10*time.Millisecond),
		WithMaxDelay(15*time.Millisecond),
		WithMultiplier(100.0),
	)
	_ = n.Send(context.Background(), makeBackoffEvents())
	elapsed := time.Since(start)
	// With max 15ms and 2 waits, should be well under 200ms.
	assert.Less(t, elapsed, 200*time.Millisecond)
}
