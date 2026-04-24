package alert_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeBurstEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: alert.EventAdded,
		Entry: scanner.Entry{Port: port, Protocol: "tcp", Address: "0.0.0.0"},
	}
}

func TestBurstNotifier_AllowsBurstImmediately(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	now := time.Now()
	clock := func() time.Time { return now }

	n := alert.NewBurstNotifier(sink, 3, 500*time.Millisecond, 5*time.Second,
		alert.WithBurstClock(clock))

	events := []alert.Event{makeBurstEvent(80), makeBurstEvent(443), makeBurstEvent(8080)}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 3 {
		t.Errorf("expected 3 events, got %d", len(received))
	}
}

func TestBurstNotifier_SuppressesAfterBurstExhausted(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	now := time.Now()
	clock := func() time.Time { return now }

	n := alert.NewBurstNotifier(sink, 2, 10*time.Second, 30*time.Second,
		alert.WithBurstClock(clock))

	// Exhaust burst.
	_ = n.Send(context.Background(), []alert.Event{makeBurstEvent(80), makeBurstEvent(443)})
	received = nil

	// Next send within rate window — should be suppressed.
	_ = n.Send(context.Background(), []alert.Event{makeBurstEvent(8080)})
	if len(received) != 0 {
		t.Errorf("expected 0 events after burst exhausted, got %d", len(received))
	}
}

func TestBurstNotifier_RefillsAfterWindow(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	now := time.Now()
	clock := func() time.Time { return now }

	n := alert.NewBurstNotifier(sink, 2, 10*time.Second, 5*time.Second,
		alert.WithBurstClock(clock))

	// Exhaust burst.
	_ = n.Send(context.Background(), []alert.Event{makeBurstEvent(80), makeBurstEvent(443)})
	received = nil

	// Advance past refill window.
	now = now.Add(6 * time.Second)
	_ = n.Send(context.Background(), []alert.Event{makeBurstEvent(8080)})
	if len(received) != 1 {
		t.Errorf("expected 1 event after refill, got %d", len(received))
	}
}

func TestBurstNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	sink := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		called = true
		return nil
	})
	n := alert.NewBurstNotifier(sink, 5, time.Second, 10*time.Second)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("inner notifier should not be called for empty events")
	}
}
