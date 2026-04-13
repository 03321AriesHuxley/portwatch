package alert_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

type countingNotifier struct {
	calls atomic.Int32
}

func (c *countingNotifier) Send(_ context.Context, events []alert.Event) error {
	if len(events) > 0 {
		c.calls.Add(1)
	}
	return nil
}

func sampleThrottleEvents() []alert.Event {
	return []alert.Event{{Kind: alert.EventAdded}}
}

func TestThrottleNotifier_AllowsFirstSend(t *testing.T) {
	inner := &countingNotifier{}
	th := alert.NewThrottleNotifier(inner, 100*time.Millisecond)

	if err := th.Send(context.Background(), sampleThrottleEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestThrottleNotifier_SuppressesWithinWindow(t *testing.T) {
	inner := &countingNotifier{}
	th := alert.NewThrottleNotifier(inner, 200*time.Millisecond)

	_ = th.Send(context.Background(), sampleThrottleEvents())
	_ = th.Send(context.Background(), sampleThrottleEvents())
	_ = th.Send(context.Background(), sampleThrottleEvents())

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestThrottleNotifier_AllowsAfterWindow(t *testing.T) {
	inner := &countingNotifier{}
	th := alert.NewThrottleNotifier(inner, 30*time.Millisecond)

	_ = th.Send(context.Background(), sampleThrottleEvents())
	time.Sleep(50 * time.Millisecond)
	_ = th.Send(context.Background(), sampleThrottleEvents())

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", inner.calls.Load())
	}
}

func TestThrottleNotifier_Reset(t *testing.T) {
	inner := &countingNotifier{}
	th := alert.NewThrottleNotifier(inner, 10*time.Second)

	_ = th.Send(context.Background(), sampleThrottleEvents())
	th.Reset()
	_ = th.Send(context.Background(), sampleThrottleEvents())

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", inner.calls.Load())
	}
}

func TestThrottleNotifier_EmptyEventsSkipped(t *testing.T) {
	inner := &countingNotifier{}
	th := alert.NewThrottleNotifier(inner, 10*time.Millisecond)

	_ = th.Send(context.Background(), []alert.Event{})
	if inner.calls.Load() != 0 {
		t.Fatalf("expected 0 calls for empty events, got %d", inner.calls.Load())
	}
}
