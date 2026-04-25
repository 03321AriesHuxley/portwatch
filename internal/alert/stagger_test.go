package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeStaggerEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: alert.EventAdded,
		Entry: scanner.Entry{
			Address:  "0.0.0.0",
			Port:     port,
			Protocol: "tcp",
		},
	}
}

func TestStaggerNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	s := alert.NewStaggerNotifier(inner, 10*time.Millisecond)
	if err := s.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("inner should not be called for empty events")
	}
}

func TestStaggerNotifier_ForwardsAllEvents(t *testing.T) {
	var received []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})
	events := []alert.Event{makeStaggerEvent(80), makeStaggerEvent(443), makeStaggerEvent(8080)}
	s := alert.NewStaggerNotifier(inner, time.Millisecond)
	if err := s.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 3 {
		t.Fatalf("expected 3 events forwarded, got %d", len(received))
	}
}

func TestStaggerNotifier_RespectsContextCancel(t *testing.T) {
	var count int
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		count++
		return nil
	})
	events := []alert.Event{makeStaggerEvent(80), makeStaggerEvent(443), makeStaggerEvent(8080)}
	s := alert.NewStaggerNotifier(inner, 100*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := s.Send(ctx, events)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 event sent before cancel, got %d", count)
	}
}

func TestStaggerNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return sentinel
	})
	s := alert.NewStaggerNotifier(inner, time.Millisecond)
	err := s.Send(context.Background(), []alert.Event{makeStaggerEvent(22)})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestStaggerNotifier_ZeroDelayClampedToMin(t *testing.T) {
	// Should not panic or busy-loop; just verifies construction and single send.
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error { return nil })
	s := alert.NewStaggerNotifier(inner, 0)
	if err := s.Send(context.Background(), []alert.Event{makeStaggerEvent(9000)}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
