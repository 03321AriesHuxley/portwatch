package alert

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func makeHedgeEvents() []Event {
	return []Event{
		{Kind: "added", Entry: makeEntry(8080, "tcp", "0.0.0.0")},
	}
}

func TestHedgeNotifier_PrimarySucceedsBeforeDelay(t *testing.T) {
	var primaryCalled atomic.Int32
	var secondaryCalled atomic.Int32

	primary := NotifierFunc(func(ctx context.Context, events []Event) error {
		primaryCalled.Add(1)
		return nil
	})
	secondary := NotifierFunc(func(ctx context.Context, events []Event) error {
		secondaryCalled.Add(1)
		return nil
	})

	h := NewHedgeNotifier(primary, secondary, 50*time.Millisecond)
	err := h.Send(context.Background(), makeHedgeEvents())

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if primaryCalled.Load() != 1 {
		t.Errorf("expected primary called once, got %d", primaryCalled.Load())
	}
	// Secondary should not have been triggered yet.
	if secondaryCalled.Load() != 0 {
		t.Errorf("expected secondary not called, got %d", secondaryCalled.Load())
	}
}

func TestHedgeNotifier_SecondaryTriggeredAfterDelay(t *testing.T) {
	var secondaryCalled atomic.Int32

	primary := NotifierFunc(func(ctx context.Context, events []Event) error {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
		}
		return ctx.Err()
	})
	secondary := NotifierFunc(func(ctx context.Context, events []Event) error {
		secondaryCalled.Add(1)
		return nil
	})

	h := NewHedgeNotifier(primary, secondary, 20*time.Millisecond)
	err := h.Send(context.Background(), makeHedgeEvents())

	if err != nil {
		t.Fatalf("expected nil error from secondary, got %v", err)
	}
	if secondaryCalled.Load() != 1 {
		t.Errorf("expected secondary called once, got %d", secondaryCalled.Load())
	}
}

func TestHedgeNotifier_BothFail_ReturnsPrimaryError(t *testing.T) {
	primaryErr := errors.New("primary failure")

	primary := NotifierFunc(func(ctx context.Context, events []Event) error {
		return primaryErr
	})
	secondary := NotifierFunc(func(ctx context.Context, events []Event) error {
		return errors.New("secondary failure")
	})

	h := NewHedgeNotifier(primary, secondary, 5*time.Millisecond)
	err := h.Send(context.Background(), makeHedgeEvents())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHedgeNotifier_EmptyEventsSkipped(t *testing.T) {
	var called atomic.Int32
	n := NotifierFunc(func(ctx context.Context, events []Event) error {
		called.Add(1)
		return nil
	})
	h := NewHedgeNotifier(n, n, 10*time.Millisecond)
	if err := h.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() != 0 {
		t.Errorf("expected no calls for empty events, got %d", called.Load())
	}
}
