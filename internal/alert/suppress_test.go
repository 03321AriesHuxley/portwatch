package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/portwatch/internal/alert"
	"github.com/yourusername/portwatch/internal/scanner"
)

func makeSuppressEvents() []alert.Event {
	return []alert.Event{
		{
			Kind: alert.EventAdded,
			Entry: scanner.Entry{Addr: "0.0.0.0", Port: 9090, Protocol: "tcp"},
		},
	}
}

func TestSuppressNotifier_ForwardsOutsideWindow(t *testing.T) {
	var received []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	sn := alert.NewSuppressNotifier(inner)
	// window in the past — should not suppress
	sn.AddWindow(time.Now().Add(-2*time.Hour), time.Now().Add(-1*time.Hour))

	if err := sn.Send(context.Background(), makeSuppressEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 event forwarded, got %d", len(received))
	}
}

func TestSuppressNotifier_SuppressesWithinWindow(t *testing.T) {
	var received []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	sn := alert.NewSuppressNotifier(inner)
	sn.AddWindow(time.Now().Add(-1*time.Minute), time.Now().Add(1*time.Hour))

	if err := sn.Send(context.Background(), makeSuppressEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 0 {
		t.Fatalf("expected 0 events (suppressed), got %d", len(received))
	}
}

func TestSuppressNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		called = true
		return nil
	})
	sn := alert.NewSuppressNotifier(inner)
	if err := sn.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("inner should not be called for empty events")
	}
}

func TestSuppressNotifier_ClearWindowsAllowsForwarding(t *testing.T) {
	var received []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	sn := alert.NewSuppressNotifier(inner)
	sn.AddWindow(time.Now().Add(-1*time.Minute), time.Now().Add(1*time.Hour))
	sn.ClearWindows()

	if err := sn.Send(context.Background(), makeSuppressEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 event after clearing windows, got %d", len(received))
	}
}

func TestSuppressNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return sentinel
	})

	sn := alert.NewSuppressNotifier(inner)
	err := sn.Send(context.Background(), makeSuppressEvents())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
