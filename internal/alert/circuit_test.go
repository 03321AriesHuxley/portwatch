package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

type countingNotifier struct {
	calls int
	err   error
}

func (c *countingNotifier) Send(_ context.Context, events []Event) error {
	c.calls++
	return c.err
}

func makeCBEvents() []Event {
	return []Event{{Kind: EventAdded}}
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	inner := &countingNotifier{}
	cb := NewCircuitBreaker(inner, 3, time.Minute)

	if err := cb.Send(context.Background(), makeCBEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected CircuitClosed, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 3, time.Minute)

	for i := 0; i < 3; i++ {
		_ = cb.Send(context.Background(), makeCBEvents())
	}

	if cb.State() != CircuitOpen {
		t.Errorf("expected CircuitOpen after %d failures, got %v", 3, cb.State())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 2, time.Hour)

	_ = cb.Send(context.Background(), makeCBEvents())
	_ = cb.Send(context.Background(), makeCBEvents())

	callsBefore := inner.calls
	err := cb.Send(context.Background(), makeCBEvents())
	if err == nil {
		t.Fatal("expected error when circuit is open")
	}
	if inner.calls != callsBefore {
		t.Errorf("inner notifier should not be called when circuit is open")
	}
}

func TestCircuitBreaker_ResetsAfterCooldown(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	now := time.Now()
	cb := NewCircuitBreaker(inner, 2, 50*time.Millisecond)
	cb.now = func() time.Time { return now }

	_ = cb.Send(context.Background(), makeCBEvents())
	_ = cb.Send(context.Background(), makeCBEvents())

	if cb.State() != CircuitOpen {
		t.Fatal("expected circuit to be open")
	}

	// advance time past cooldown
	cb.now = func() time.Time { return now.Add(100 * time.Millisecond) }
	inner.err = nil // downstream recovers

	if err := cb.Send(context.Background(), makeCBEvents()); err != nil {
		t.Fatalf("expected success after cooldown, got: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected CircuitClosed after recovery, got %v", cb.State())
	}
}
