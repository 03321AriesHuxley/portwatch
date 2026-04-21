package alert_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/portwatch/internal/alert"
)

func makeEscalateEvents() []alert.Event {
	return []alert.Event{
		{Kind: "added", Entry: makeEntry(8080, "tcp", "0.0.0.0")},
	}
}

func TestEscalateNotifier_PrimarySucceeds(t *testing.T) {
	var called int32
	primary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	n := alert.NewEscalateNotifier(primary)
	if err := n.Send(context.Background(), makeEscalateEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected primary called once, got %d", called)
	}
}

func TestEscalateNotifier_EscalatesAfterDelay(t *testing.T) {
	primaryErr := errors.New("primary down")
	primary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return primaryErr
	})

	var escalated int32
	secondary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		atomic.AddInt32(&escalated, 1)
		return nil
	})

	n := alert.NewEscalateNotifier(primary,
		alert.EscalationPolicy{After: 0, Target: secondary},
	)

	if err := n.Send(context.Background(), makeEscalateEvents()); err != nil {
		t.Fatalf("expected escalation to succeed, got: %v", err)
	}
	if atomic.LoadInt32(&escalated) != 1 {
		t.Fatalf("expected secondary called once, got %d", escalated)
	}
}

func TestEscalateNotifier_DoesNotEscalateBeforeDelay(t *testing.T) {
	primaryErr := errors.New("primary down")
	primary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return primaryErr
	})

	var escalated int32
	secondary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		atomic.AddInt32(&escalated, 1)
		return nil
	})

	n := alert.NewEscalateNotifier(primary,
		alert.EscalationPolicy{After: 10 * time.Minute, Target: secondary},
	)

	err := n.Send(context.Background(), makeEscalateEvents())
	if err == nil {
		t.Fatal("expected error when escalation window not reached")
	}
	if atomic.LoadInt32(&escalated) != 0 {
		t.Fatalf("expected secondary not called, got %d", escalated)
	}
}

func TestEscalateNotifier_ResetOnSuccess(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	primary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		if fail.Load() {
			return errors.New("down")
		}
		return nil
	})

	var escalated int32
	secondary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		atomic.AddInt32(&escalated, 1)
		return nil
	})

	n := alert.NewEscalateNotifier(primary,
		alert.EscalationPolicy{After: 0, Target: secondary},
	)

	_ = n.Send(context.Background(), makeEscalateEvents())
	fail.Store(false)
	if err := n.Send(context.Background(), makeEscalateEvents()); err != nil {
		t.Fatalf("expected success after recovery: %v", err)
	}
	if atomic.LoadInt32(&escalated) != 1 {
		t.Fatalf("expected secondary called once, got %d", escalated)
	}
}
