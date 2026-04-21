package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/portwatch/internal/alert"
)

func makeMuteEvents() []alert.Event {
	return []alert.Event{
		{Kind: alert.EventAdded, Port: 8080, Proto: "tcp", Addr: "0.0.0.0"},
	}
}

func TestMuteNotifier_ForwardsWhenNotMuted(t *testing.T) {
	var called bool
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	mn := alert.NewMuteNotifier(inner)

	if err := mn.Send(context.Background(), makeMuteEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected inner notifier to be called")
	}
}

func TestMuteNotifier_SuppressesWhenMuted(t *testing.T) {
	var called bool
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	mn := alert.NewMuteNotifier(inner)
	mn.Mute(5 * time.Second)

	if err := mn.Send(context.Background(), makeMuteEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected inner notifier NOT to be called while muted")
	}
}

func TestMuteNotifier_AllowsAfterUnmute(t *testing.T) {
	var called bool
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	mn := alert.NewMuteNotifier(inner)
	mn.Mute(5 * time.Second)
	mn.Unmute()

	if err := mn.Send(context.Background(), makeMuteEvents()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected inner notifier to be called after unmute")
	}
}

func TestMuteNotifier_EmptyEventsSkipped(t *testing.T) {
	var called bool
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	mn := alert.NewMuteNotifier(inner)

	if err := mn.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected inner notifier NOT to be called for empty events")
	}
}

func TestMuteNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		return sentinel
	})
	mn := alert.NewMuteNotifier(inner)

	err := mn.Send(context.Background(), makeMuteEvents())
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestMuteNotifier_IsMuted(t *testing.T) {
	mn := alert.NewMuteNotifier(alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error { return nil }))

	if mn.IsMuted() {
		t.Error("should not be muted initially")
	}
	mn.Mute(10 * time.Second)
	if !mn.IsMuted() {
		t.Error("should be muted after Mute()")
	}
	mn.Unmute()
	if mn.IsMuted() {
		t.Error("should not be muted after Unmute()")
	}
}
