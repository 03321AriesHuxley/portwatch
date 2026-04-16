package alert_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/portwatch/internal/alert"
)

func makeCooldownEvents() []alert.Event {
	return []alert.Event{{Kind: alert.EventAdded}}
}

func TestCooldownNotifier_AllowsBelowThreshold(t *testing.T) {
	var sent int
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		sent += len(events)
		return nil
	})

	cn := alert.NewCooldownNotifier(inner, 3, time.Second, 5*time.Second)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = cn.Send(ctx, makeCooldownEvents())
	}

	if sent != 3 {
		t.Fatalf("expected 3 sent, got %d", sent)
	}
}

func TestCooldownNotifier_SuppressesAfterBurst(t *testing.T) {
	var sent int
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		sent += len(events)
		return nil
	})

	cn := alert.NewCooldownNotifier(inner, 2, time.Second, 10*time.Second)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = cn.Send(ctx, makeCooldownEvents())
	}

	// first 2 pass, 3rd triggers cooldown (len becomes 3 > threshold 2)
	if sent > 2 {
		t.Fatalf("expected at most 2 sent during cooldown, got %d", sent)
	}
}

func TestCooldownNotifier_AllowsAfterCooldown(t *testing.T) {
	var sent int
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		sent += len(events)
		return nil
	})

	cn := alert.NewCooldownNotifier(inner, 1, 50*time.Millisecond, 80*time.Millisecond)
	ctx := context.Background()

	_ = cn.Send(ctx, makeCooldownEvents())
	_ = cn.Send(ctx, makeCooldownEvents()) // triggers cooldown

	time.Sleep(120 * time.Millisecond)

	pre := sent
	_ = cn.Send(ctx, makeCooldownEvents())

	if sent == pre {
		t.Fatal("expected event to be forwarded after cooldown expired")
	}
}

func TestCooldownNotifier_EmptyEventsSkipped(t *testing.T) {
	var called bool
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		called = true
		return nil
	})

	cn := alert.NewCooldownNotifier(inner, 5, time.Second, time.Second)
	_ = cn.Send(context.Background(), nil)

	if called {
		t.Fatal("expected inner not to be called for empty events")
	}
}
