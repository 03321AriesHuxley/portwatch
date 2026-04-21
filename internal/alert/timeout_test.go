package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeTimeoutEvent() alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{Addr: "0.0.0.0", Port: 9090, Protocol: "tcp"},
	}
}

func TestTimeoutNotifier_SucceedsWithinDeadline(t *testing.T) {
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	tn := alert.NewTimeoutNotifier(inner, 200*time.Millisecond)

	if err := tn.Send(context.Background(), []alert.Event{makeTimeoutEvent()}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestTimeoutNotifier_ExceedsDeadline(t *testing.T) {
	inner := alert.NotifierFunc(func(ctx context.Context, _ []alert.Event) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	tn := alert.NewTimeoutNotifier(inner, 30*time.Millisecond)

	err := tn.Send(context.Background(), []alert.Event{makeTimeoutEvent()})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded in error chain, got: %v", err)
	}
}

func TestTimeoutNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		called = true
		return nil
	})
	tn := alert.NewTimeoutNotifier(inner, 100*time.Millisecond)

	if err := tn.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("inner notifier should not be called for empty events")
	}
}

func TestTimeoutNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return sentinel
	})
	tn := alert.NewTimeoutNotifier(inner, 200*time.Millisecond)

	err := tn.Send(context.Background(), []alert.Event{makeTimeoutEvent()})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}

func TestTimeoutNotifier_DefaultsToFiveSeconds(t *testing.T) {
	// Passing zero or negative timeout should default to 5s (no panic, no zero deadline).
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return nil
	})
	tn := alert.NewTimeoutNotifier(inner, 0)
	if tn == nil {
		t.Fatal("expected non-nil TimeoutNotifier")
	}
	if err := tn.Send(context.Background(), []alert.Event{makeTimeoutEvent()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
