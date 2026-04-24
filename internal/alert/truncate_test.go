package alert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeTruncateEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{
			LocalAddress: "0.0.0.0",
			LocalPort:    port,
			Protocol:     "tcp",
		},
	}
}

func TestTruncateNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	fn := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	n := alert.NewTruncateNotifier(fn, 5)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("inner should not be called for empty events")
	}
}

func TestTruncateNotifier_ForwardsWithinLimit(t *testing.T) {
	var got []alert.Event
	fn := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewTruncateNotifier(fn, 5)
	events := []alert.Event{
		makeTruncateEvent(80),
		makeTruncateEvent(443),
		makeTruncateEvent(8080),
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
}

func TestTruncateNotifier_TruncatesAboveLimit(t *testing.T) {
	var got []alert.Event
	fn := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewTruncateNotifier(fn, 2)
	events := []alert.Event{
		makeTruncateEvent(80),
		makeTruncateEvent(443),
		makeTruncateEvent(8080),
		makeTruncateEvent(9090),
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events after truncation, got %d", len(got))
	}
	if got[0].Entry.LocalPort != 80 || got[1].Entry.LocalPort != 443 {
		t.Fatalf("expected first two events to be forwarded, got ports %d %d",
			got[0].Entry.LocalPort, got[1].Entry.LocalPort)
	}
}

func TestTruncateNotifier_ZeroMaxDefaultsToOne(t *testing.T) {
	var got []alert.Event
	fn := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewTruncateNotifier(fn, 0)
	events := []alert.Event{makeTruncateEvent(80), makeTruncateEvent(443)}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event (clamped from 0), got %d", len(got))
	}
}

func TestTruncateNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	fn := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return sentinel
	})
	n := alert.NewTruncateNotifier(fn, 10)
	err := n.Send(context.Background(), []alert.Event{makeTruncateEvent(80)})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
