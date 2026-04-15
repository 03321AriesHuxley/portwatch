package alert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeJitterEvents() []Event {
	return []Event{{Kind: EventAdded, Port: 9090, Proto: "tcp", Addr: "127.0.0.1"}}
}

func TestJitterNotifier_EventuallyDelegates(t *testing.T) {
	var called bool
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			called = true
			return nil
		},
	}
	n := NewJitterNotifier(inner, 1*time.Millisecond, 5*time.Millisecond)
	err := n.Send(context.Background(), makeJitterEvents())
	require.NoError(t, err)
	assert.True(t, called)
}

func TestJitterNotifier_RespectsContextCancel(t *testing.T) {
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	n := NewJitterNotifier(inner, 500*time.Millisecond, 1*time.Second)
	err := n.Send(ctx, makeJitterEvents())
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestJitterNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner error")
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			return sentinel
		},
	}
	n := NewJitterNotifier(inner, 1*time.Millisecond, 2*time.Millisecond)
	err := n.Send(context.Background(), makeJitterEvents())
	require.ErrorIs(t, err, sentinel)
}

func TestJitterNotifier_MaxDelayClampedWhenEqual(t *testing.T) {
	// When min == max, constructor should not panic and should still work.
	var called bool
	inner := &mockNotifier{
		sendFn: func(_ context.Context, _ []Event) error {
			called = true
			return nil
		},
	}
	n := NewJitterNotifier(inner, 5*time.Millisecond, 5*time.Millisecond)
	err := n.Send(context.Background(), makeJitterEvents())
	require.NoError(t, err)
	assert.True(t, called)
}
